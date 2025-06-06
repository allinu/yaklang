package wordparser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils/fileparser/types"
)

// FileType 定义文件类型
type FileType string

const (
	FileTypeText      FileType = "text"
	FileTypeTable     FileType = "table"
	FileTypeImage     FileType = "image"
	FileTypeChart     FileType = "chart"
	FileTypePDF       FileType = "pdf"
	FileTypeOLE       FileType = "ole"
	FileTypeVBA       FileType = "vba"
	FileTypeVideo     FileType = "video"
	FileTypeVideoData FileType = "videodata" // 视频二进制数据
)

// File 定义文件对象
type File struct {
	Name       string            // 文件名
	Data       []byte            // 文件内容（兼容旧版本）
	Content    string            // 文本内容
	BinaryData []byte            // 二进制数据
	FileName   string            // 文件名（用于二进制文件）
	Metadata   map[string]string // 附加信息
}

// DumpToFiles 将 WordNodeClassifier 对象转换为文件对象
func (c *WordNodeClassifier) DumpToFiles() map[string][]types.File {
	result := make(map[string][]types.File)

	// 处理文本内容
	if len(c.Texts) > 0 {
		var textBuffer bytes.Buffer
		for _, text := range c.Texts {
			// 根据文本样式添加标记
			if text.IsBold {
				textBuffer.WriteString("**")
			}
			if text.IsItalic {
				textBuffer.WriteString("*")
			}
			if text.IsStrike {
				textBuffer.WriteString("~~")
			}
			textBuffer.WriteString(text.Text)
			if text.IsStrike {
				textBuffer.WriteString("~~")
			}
			if text.IsItalic {
				textBuffer.WriteString("*")
			}
			if text.IsBold {
				textBuffer.WriteString("**")
			}
			textBuffer.WriteString("\n")
		}
		result[string(FileTypeText)] = []types.File{
			{
				FileName:   "text/content.txt",
				BinaryData: textBuffer.Bytes(),
				Metadata: map[string]string{
					"count": fmt.Sprintf("%d", len(c.Texts)),
				},
			},
		}
		log.Debugf("已导出文本内容到 content.txt，共 %d 段", len(c.Texts))
	}

	// 处理表格内容
	if len(c.Tables) > 0 {
		var tableBuffer bytes.Buffer
		for i, table := range c.Tables {
			// 写入表格标题
			tableBuffer.WriteString(fmt.Sprintf("## 表格 %d\n\n", i+1))

			// 写入表头
			tableBuffer.WriteString("| ")
			tableBuffer.WriteString(strings.Join(table.Headers, " | "))
			tableBuffer.WriteString(" |\n")

			// 写入分隔行
			tableBuffer.WriteString("|")
			for range table.Headers {
				tableBuffer.WriteString(" --- |")
			}
			tableBuffer.WriteString("\n")

			// 写入数据行
			for _, row := range table.Rows {
				tableBuffer.WriteString("| ")
				tableBuffer.WriteString(strings.Join(row, " | "))
				tableBuffer.WriteString(" |\n")
			}
			tableBuffer.WriteString("\n")
		}
		result[string(FileTypeTable)] = []types.File{
			{
				FileName:   "tables/tables.md",
				BinaryData: tableBuffer.Bytes(),
				Metadata: map[string]string{
					"count": fmt.Sprintf("%d", len(c.Tables)),
				},
			},
		}
		log.Debugf("已导出表格内容到 tables.md，共 %d 个表格", len(c.Tables))
	}

	// 处理图片内容
	if len(c.Images) > 0 {
		var images []types.File
		for i, img := range c.Images {
			ext := ".bin"
			switch img.MimeType {
			case "image/png":
				ext = ".png"
			case "image/jpeg":
				ext = ".jpg"
			}

			filename := img.Name
			if filename == "" {
				filename = fmt.Sprintf("image_%d%s", i+1, ext)
			}

			images = append(images, types.File{
				FileName:   "images/" + filename,
				BinaryData: img.Data,
				Metadata: map[string]string{
					"mime_type": img.MimeType,
					"size":      fmt.Sprintf("%d", len(img.Data)),
				},
			})
			log.Debugf("已导出图片 %s，大小: %d bytes", filename, len(img.Data))
		}
		result[string(FileTypeImage)] = images
	}

	// 处理图表内容
	if len(c.Charts) > 0 {
		var charts []types.File
		for i, chart := range c.Charts {
			filename := fmt.Sprintf("chart_%d.xml", i+1)
			charts = append(charts, types.File{
				FileName:   "charts/" + filename,
				BinaryData: chart.Data,
				Metadata: map[string]string{
					"type": chart.Type,
					"size": fmt.Sprintf("%d", len(chart.Data)),
				},
			})
			log.Debugf("已导出图表 %s，类型: %s", filename, chart.Type)
		}
		result[string(FileTypeChart)] = charts
	}

	// 处理PDF附件
	if len(c.PDFs) > 0 {
		var pdfs []types.File
		for i, pdf := range c.PDFs {
			filename := pdf.Name
			if filename == "" {
				filename = fmt.Sprintf("attachment_%d.pdf", i+1)
			}
			pdfs = append(pdfs, types.File{
				FileName:   "pdfs/" + filename,
				BinaryData: pdf.Data,
				Metadata: map[string]string{
					"size": fmt.Sprintf("%d", len(pdf.Data)),
				},
			})
			log.Debugf("已导出PDF附件 %s，大小: %d bytes", filename, len(pdf.Data))
		}
		result[string(FileTypePDF)] = pdfs
	}

	// 处理OLE对象
	if len(c.OLEs) > 0 {
		var oles []types.File
		for i, ole := range c.OLEs {
			filename := ole.Name
			if filename == "" {
				filename = fmt.Sprintf("ole_%d.bin", i+1)
			}
			oles = append(oles, types.File{
				FileName:   "oles/" + filename,
				BinaryData: ole.Data,
				Metadata: map[string]string{
					"type": ole.Type,
					"size": fmt.Sprintf("%d", len(ole.Data)),
				},
			})
			log.Debugf("已导出OLE对象 %s，类型: %s", filename, ole.Type)
		}
		result[string(FileTypeOLE)] = oles
	}

	// 处理VBA代码
	if len(c.VBAs) > 0 {
		var vbas []types.File
		for i, vba := range c.VBAs {
			filename := fmt.Sprintf("%s.vba", vba.ModName)
			if vba.ModName == "Unknown" {
				filename = fmt.Sprintf("macro_%d.vba", i+1)
			}
			vbas = append(vbas, types.File{
				FileName:   "vbas/" + filename,
				BinaryData: []byte(vba.Code),
				Metadata: map[string]string{
					"module": vba.ModName,
					"size":   fmt.Sprintf("%d", len(vba.Code)),
				},
			})
			log.Debugf("已导出VBA代码 %s，模块: %s", filename, vba.ModName)
		}
		result[string(FileTypeVBA)] = vbas
	}

	return result
}

// GetFileExtension 根据MIME类型获取文件扩展名
func GetFileExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	case "text/markdown":
		return ".md"
	default:
		return ".bin"
	}
}
