package yakit

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/schema"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bizhelper"
	"github.com/yaklang/yaklang/common/yak/ssa/ssadb"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

func FilterSSAProgram(db *gorm.DB, filter *ypb.SSAProgramFilter) *gorm.DB {
	db = db.Model(&ssadb.IrProgram{})
	if filter == nil {
		return db
	}

	db = bizhelper.ExactOrQueryStringArrayOr(db, "program_name", filter.GetProgramNames())
	db = bizhelper.ExactOrQueryStringArrayOr(db, "language", filter.GetLanguages())
	db = bizhelper.ExactQueryInt64ArrayOr(db, "id", filter.GetIds())

	if word := filter.GetKeyword(); word != "" {
		db = bizhelper.FuzzSearchEx(db, []string{"program_name", "description"}, word, false)
	}
	if filter.GetAfterID() > 0 {
		db = db.Where("id > ?", filter.GetAfterID())
	}
	if filter.GetBeforeID() > 0 {
		db = db.Where("id < ?", filter.GetBeforeID())
	}

	if filter.GetBeforeUpdatedAt() > 0 {
		db = bizhelper.QueryByTimeRangeWithTimestamp(db, "updated_at", 0, filter.GetBeforeUpdatedAt())
	}
	if filter.GetAfterUpdatedAt() > 0 {
		db = bizhelper.QueryByTimeRangeWithTimestamp(db, "updated_at", filter.GetAfterUpdatedAt(), time.Now().Add(10*time.Minute).Unix())
	}
	return db
}

// when filter == nil delete all program
func DeleteSSAProgram(DB *gorm.DB, filter *ypb.SSAProgramFilter) (int, error) {
	// db is profile database
	db := DB.Model(&ssadb.IrProgram{})
	db = FilterSSAProgram(db, filter)
	// get all program
	var programs []*ssadb.IrProgram
	queryResult := db.Model(&ssadb.IrProgram{}).Select("program_name").Find(&programs)
	if queryResult.Error != nil {
		log.Errorf("query ssa program fail: %s", queryResult.Error)
	}
	// delete schema program
	result := db.Unscoped().Delete(&ssadb.IrProgram{})

	// delete ssadb program
	programNames := make([]string, 0, len(programs))
	for _, prog := range programs {
		ssadb.DeleteProgram(ssadb.GetDB(), prog.ProgramName)
		programNames = append(programNames, prog.ProgramName)
	}
	// delete risk create by this program
	DeleteSSARisks(ssadb.GetDB(), &ypb.SSARisksFilter{ProgramName: programNames})
	// DeleteRisk()
	return len(programs), result.Error
}

func QuerySSAProgram(db *gorm.DB, request *ypb.QuerySSAProgramRequest) (*bizhelper.Paginator, []*ypb.SSAProgram, error) {
	var programs []*ssadb.IrProgram
	p := request.Pagination
	if p == nil {
		p = &ypb.Paging{
			Page:    1,
			Limit:   30,
			OrderBy: "updated_at",
			Order:   "desc",
		}
	}
	db = bizhelper.QueryOrder(db, p.OrderBy, p.Order)
	db = FilterSSAProgram(db, request.GetFilter())
	paging, dbx := bizhelper.Paging(db, int(p.Page), int(p.Limit), &programs)
	if dbx.Error != nil {
		return nil, nil, utils.Errorf("select ssa program fail: %s", dbx.Error)
	}
	progGRPCs := make([]*ypb.SSAProgram, 0, len(programs))
	for _, prog := range programs {
		progGRPCs = append(progGRPCs, Prog2GRPC(prog))
	}
	return paging, progGRPCs, nil
}

func Prog2GRPC(prog *ssadb.IrProgram) *ypb.SSAProgram {
	ret := &ypb.SSAProgram{
		Id: uint32(prog.ID),
		// basic info
		CreateAt:      prog.CreatedAt.Unix(),
		UpdateAt:      prog.UpdatedAt.Unix(),
		Name:          prog.ProgramName,
		Description:   prog.Description,
		Language:      prog.Language,
		EngineVersion: prog.EngineVersion,
		Dbpath:        consts.GetSSADataBasePathDefault(consts.GetDefaultYakitBaseDir()),
	}
	// recompile
	NeedReCompile := func() bool {
		return false
		// return prog.EngineVersion != consts.GetYakVersion()
	}
	ret.Recompile = NeedReCompile()
	// risk info
	{
		var result struct {
			High     int64
			Low      int64
			Critical int64
			Warning  int64
		}
		projectDB := ssadb.GetDB()
		if err := projectDB.Model(&schema.SSARisk{}).Where("program_name=?", prog.ProgramName).Select(`
		sum(case when severity='critical' then 1 else 0 end) as critical,
		sum(case when severity='high' then 1 else 0 end) as high,
		sum(case when severity='warning' then 1 else 0 end) as warning,
		sum(case when severity='low' then 1 else 0 end) as low
		`).Scan(&result).Error; err != nil {
			log.Errorf("query risk fail: %s", err) // ignore
		}
		ret.CriticalRiskNumber = result.Critical
		ret.HighRiskNumber = result.High
		ret.WarnRiskNumber = result.Warning
		ret.LowRiskNumber = result.Low
	}

	return ret
}

func UpdateSSAProgram(DB *gorm.DB, input *ypb.SSAProgramInput) (int64, error) {
	if input == nil {
		return 0, utils.Errorf("input is nil ")
	}
	db := DB.Model(&ssadb.IrProgram{})
	db = db.Where("program_name = ?", input.GetName()).Update("description", input.Description)
	return db.RowsAffected, db.Error
}
