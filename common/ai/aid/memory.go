package aid

import (
	"bytes"
	"github.com/yaklang/yaklang/common/ai/aid/aitool"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils/omap"
	osRuntime "runtime"
	"strings"
	"text/template"
	"time"
)

type PlanRecord struct { // todo
	PlanRequest  *planRequest
	PlanResponse *PlanResponse
}

type InteractiveEventRecord struct {
	InteractiveEvent *Event
	UserInput        aitool.InvokeParams
}

type Memory struct {
	// user first input
	Query string

	// persistent data
	PersistentData []string

	// user data, ai or user can write and read
	userData *omap.OrderedMap[string, string]

	// task info
	CurrentTask *aiTask
	RootTask    *aiTask

	// todo
	PlanHistory []*PlanRecord

	// tools list
	DisableTools bool
	Tools        func() []*aitool.Tool

	// tool call results
	toolCallResults []*aitool.ToolResult

	// interactive history
	InteractiveHistory *omap.OrderedMap[string, *InteractiveEventRecord]
}

func NewMemory() *Memory {
	return &Memory{
		PlanHistory:        make([]*PlanRecord, 0),
		PersistentData:     make([]string, 0),
		InteractiveHistory: omap.NewOrderedMap[string, *InteractiveEventRecord](make(map[string]*InteractiveEventRecord)),
		toolCallResults:    make([]*aitool.ToolResult, 0),
		Tools: func() []*aitool.Tool {
			return make([]*aitool.Tool, 0)
		},
		userData: omap.NewOrderedMap[string, string](make(map[string]string)),
	}
}

// user data memory api, user or ai can set and get
func (m *Memory) UserDataKeys() []string {
	return m.userData.Keys()
}

func (m *Memory) UserDataGet(key string) (string, bool) {
	return m.userData.Get(key)
}

func (m *Memory) UserDataDelete(key string) {
	m.userData.Delete(key)
	return
}

func (m *Memory) StoreUserData(key string, value string) {
	m.userData.Set(key, value)
}

// constants info memory api
func (m *Memory) Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (m *Memory) OS() string {
	return osRuntime.GOOS
}

func (m *Memory) Arch() string {
	return osRuntime.GOARCH
}

func (m *Memory) Schema() map[string]string {
	return taskJSONSchema()
}

// set tools list
func (m *Memory) StoreTools(toolList func() []*aitool.Tool) {
	m.Tools = toolList
}

// user first input
func (m *Memory) StoreQuery(query string) {
	m.Query = query
}

// task info memory
func (m *Memory) StoreRootTask(t *aiTask) {
	m.RootTask = t
}

func (m *Memory) Progress() string {
	return m.RootTask.Progress()
}

func (m *Memory) StoreCurrentTask(task *aiTask) {
	m.CurrentTask = task
}

func (m *Memory) StoreAppendPersistentInfo(i ...string) {
	m.PersistentData = append(m.PersistentData, i...)
}

// interactive history memory
func (m *Memory) StoreInteractiveEvent(eventID string, e *Event) {
	m.InteractiveHistory.Set(eventID, &InteractiveEventRecord{
		InteractiveEvent: e,
	})
}

func (m *Memory) StoreInteractiveUserInput(eventID string, invoke aitool.InvokeParams) {
	record, ok := m.InteractiveHistory.Get(eventID)
	if !ok {
		log.Errorf("error getting review record for event ID %s", eventID)
		return
	}
	record.UserInput = invoke
}

// tool results memory
func (m *Memory) PushToolCallResults(t ...*aitool.ToolResult) {
	m.toolCallResults = append(m.toolCallResults, t...)
}

func (m *Memory) PromptForToolCallResultsForLastN(n int) string {
	if len(m.toolCallResults) == 0 {
		return ""
	}

	var result = m.toolCallResults
	if len(result) > n {
		result = result[len(result)-n:]
	}
	templateData := map[string]interface{}{
		"ToolCallResults": result,
	}
	temp, err := template.New("tool-result-history").Parse(__prompt_ToolResultHistoryPromptTemplate)
	if err != nil {
		log.Errorf("error parsing tool result history template: %v", err)
		return ""
	}
	var promptBuilder strings.Builder
	err = temp.Execute(&promptBuilder, templateData)
	if err != nil {
		log.Errorf("error executing tool result history template: %v", err)
		return ""
	}
	return promptBuilder.String()
}

func (m *Memory) PromptForToolCallResultsForLast5() string {
	return m.PromptForToolCallResultsForLastN(5)
}

func (m *Memory) PromptForToolCallResultsForLast10() string {
	return m.PromptForToolCallResultsForLastN(10)
}

func (m *Memory) PromptForToolCallResultsForLast20() string {
	return m.PromptForToolCallResultsForLastN(20)
}

// memory tools current task info
func (m *Memory) CurrentTaskInfo() string {
	if m.CurrentTask == nil {
		return ""
	}
	templateData := map[string]interface{}{
		"Memory": m,
	}
	temp, err := template.New("current_task_info").Parse(__prompt_currentTaskInfo)
	if err != nil {
		log.Errorf("error parsing tool result history template: %v", err)
		return ""
	}
	var promptBuilder strings.Builder
	err = temp.Execute(&promptBuilder, templateData)
	if err != nil {
		log.Errorf("error executing tool result history template: %v", err)
		return ""
	}
	return promptBuilder.String()
}

func (m *Memory) PersistentMemory() string {
	var buf bytes.Buffer
	buf.WriteString("# Now " + time.Now().String() + "\n")
	for _, info := range m.PersistentData {
		buf.WriteString(info)
		buf.WriteString("\n")
	}
	return buf.String()
}

func (m *Memory) ToolsList() string {
	if m.DisableTools {
		return ""
	}
	templateData := map[string]interface{}{
		"Tools": m.Tools(),
	}
	temp, err := template.New("tools_list").Parse(__prompt_ToolsList)
	if err != nil {
		log.Errorf("error parsing tool result history template: %v", err)
		return ""
	}
	var promptBuilder strings.Builder
	err = temp.Execute(&promptBuilder, templateData)
	if err != nil {
		log.Errorf("error executing tool result history template: %v", err)
		return ""
	}
	return promptBuilder.String()
}
