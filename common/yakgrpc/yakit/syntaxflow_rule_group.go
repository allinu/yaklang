package yakit

import (
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/schema"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bizhelper"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

type GroupAndRuleCount struct {
	GroupName string
	Count     int64
}

// QuerySyntaxFlowRuleGroup 查询规则组中相关规则的个数
func QuerySyntaxFlowRuleGroup(db *gorm.DB, params *ypb.QuerySyntaxFlowRuleGroupRequest) (result []*schema.SyntaxFlowGroup, err error) {
	if params == nil {
		return nil, utils.Error("query syntax flow rule group failed: query params is nil")
	}
	db = db.Model(&schema.SyntaxFlowGroup{}).Preload("Rules")
	db = FilterSyntaxFlowGroups(db, params.GetFilter())
	if err = db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func FilterSyntaxFlowGroups(db *gorm.DB, filter *ypb.SyntaxFlowRuleGroupFilter) *gorm.DB {
	if filter == nil {
		return db
	}
	db = bizhelper.ExactOrQueryStringArrayOr(db, "group_name", filter.GetGroupNames())
	if filter.GetKeyWord() != "" {
		db = bizhelper.FuzzQueryStringArrayOrLike(db,
			"group_name", []string{filter.GetKeyWord()})
	}
	if filter.GetFilterGroupKind() != "" {
		if filter.GetFilterGroupKind() == "buildIn" {
			db = db.Where("is_build_in = ?", true)
		} else if filter.GetFilterGroupKind() == "unBuildIn" {
			db = db.Where("is_build_in = ?", false)
		}
	}
	return db
}

func DeleteSyntaxFlowRuleGroup(db *gorm.DB, params *ypb.DeleteSyntaxFlowRuleGroupRequest) (int64, error) {
	if params == nil {
		return 0, utils.Error("delete syntax flow rule group failed: delete syntaxflow rule request is nil")
	}
	if params.GetFilter() == nil {
		return 0, utils.Error("delete syntax flow rule group failed: delete filter is nil")
	}

	db = FilterSyntaxFlowGroups(db, params.GetFilter())
	db = db.Model(&schema.SyntaxFlowGroup{}).
		Unscoped().Delete(&schema.SyntaxFlowGroup{})
	return db.RowsAffected, db.Error
}

func QuerySyntaxFlowGroupCount(db *gorm.DB, groupNames []string) int64 {
	db = db.Model(&schema.SyntaxFlowGroup{})
	var count int64
	db.Where("group_name IN (?)", groupNames).Count(&count)
	return count
}
