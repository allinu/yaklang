package cve

import (
	"context"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/ai"
	"github.com/yaklang/yaklang/common/ai/aispec"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/cve/cveresources"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

func TranslatingCWE(apiKeyFile string, concurrent int, cveResourceDb string) error {
	var keyStr string
	var err error
	if apiKeyFile != "" {
		key, err := ioutil.ReadFile(apiKeyFile)
		if err != nil {
			return err
		}
		keyStr = strings.TrimSpace(string(key))
	}
	var db *gorm.DB // = consts.GetGormCVEDatabase()
	if cveResourceDb == "" {
		db = consts.GetGormCVEDatabase()
	} else {
		db, err = consts.CreateCVEDatabase(cveResourceDb)
		if err != nil {
			log.Errorf("cannot open: %s with error: %s", cveResourceDb, err)
		}
	}
	if db == nil {
		return utils.Error("no cve database found")
	}

	descDB := consts.GetGormCVEDescriptionDatabase()
	if descDB == nil {
		return utils.Error("empty description database")
	}

	db = db.Model(&cveresources.CWE{}).Where(
		"(name_zh = '') OR " +
			"(description_zh = '') OR " +
			"(extended_description_zh = '') OR " +
			"(cwe_solution = '')")
	if concurrent <= 0 {
		concurrent = 10
	}
	var count int64
	db.Count(&count)
	if count > 0 {
		log.Infof("rest total: %v", count)
	}
	for r := range cveresources.YieldCWEs(db, context.Background()) {
		cveresources.CreateOrUpdateCWE(descDB, r.IdStr, r)
	}
	swg := utils.NewSizedWaitGroup(concurrent)
	current := 0
	for c := range cveresources.YieldCWEs(descDB, context.Background()) {
		current++
		c := c
		swg.Add()
		go func() {
			defer func() {
				swg.Done()
			}()
			start := time.Now()
			cweIns, err := MakeOpenAITranslateCWE(c, keyStr)
			log.Infof(
				"%6d/%-6d save [%v] chinese desc finished: cost: %v",
				current, count, c.CWEString(), time.Now().Sub(start).String(),
			)
			if err != nil {
				if !strings.Contains(err.Error(), `translating existed`) {
					log.Errorf("make openai working failed cwe: %s", err)
				}

				if strings.Contains(err.Error(), `Service Unavailable`) {
					time.Sleep(time.Minute)
				}
				return
			}
			cveresources.CreateOrUpdateCWE(descDB, cweIns.IdStr, cweIns)
			end := time.Now()
			if dur := end.Sub(start); dur.Seconds() > 3 {
				return
			} else {
				time.Sleep(time.Duration(math.Floor(float64(3)-dur.Seconds())+1) * time.Second)
			}
		}()
	}
	swg.Wait()
	return nil
}

func Translating(aiGatewayType string, noCritical bool, concurrent int, cveResourceDb string, opts ...aispec.AIConfigOption) error {
	var err error

	if aiGatewayType == "" {
		aiGatewayType = "openai"
	}

	if !ai.HaveAI(aiGatewayType) {
		return utils.Errorf("ai gateway type: %s not found", aiGatewayType)
	}

	var db *gorm.DB // = consts.GetGormCVEDatabase()
	if cveResourceDb == "" {
		db = consts.GetGormCVEDatabase()
	} else {
		db, err = consts.CreateCVEDatabase(cveResourceDb)
		if err != nil {
			log.Errorf("cannot open: %s with error: %s", cveResourceDb, err)
		}
	}
	if db == nil {
		return utils.Error("no cve database found")
	}

	db = db.Model(&cveresources.CVE{}).Where("(title_zh is null) OR (title_zh = '')")
	if concurrent <= 0 {
		concurrent = 10
	}

	if os.Getenv("ASC") != "" {
		db = db.Order("published_date asc")
	} else {
		db = db.Order("published_date desc")
	}
	var count int64
	db.Count(&count)
	if count > 0 {
		log.Infof("rest total: %v", count)
	}
	swg := utils.NewSizedWaitGroup(concurrent)
	var current int64 = 0

	aiClient := ai.GetAI(aiGatewayType, opts...)

	for c := range cveresources.YieldCVEs(db, context.Background()) {
		atomic.AddInt64(&current, 1)

		lowlevel := c.BaseCVSSv2Score <= 6.0 && c.ImpactScore <= 6.0 && c.ExploitabilityScore <= 6.0
		if !((lowlevel && noCritical) || (!lowlevel && !noCritical)) {
			continue
		}
		c := c
		swg.Add()
		go func(idx int64) {
			defer func() {
				swg.Done()
			}()
			start := time.Now()

			err := MakeOpenAIWorking(c, aiClient)
			cost := time.Now().Sub(start)
			if cost.Seconds() > 2 {
				log.Infof(
					"%6d/%-6d save [%v] chinese cve desc finished: cost: %v",
					idx, count, c.CVE, cost.String(),
				)
			} else {
				log.Infof(
					"%6d/%-6d save [%v] chinese cve desc finished: cost: %v: %v",
					idx, count, c.CVE, cost.String(),
					c.TitleZh,
				)
			}
			if err != nil {
				if !strings.Contains(err.Error(), `translating existed`) {
					log.Errorf("make openai working failed: %s", err)
				}

				if strings.Contains(err.Error(), `Service Unavailable`) {
					time.Sleep(time.Minute)
				}
				return
			}
			end := time.Now()
			if dur := end.Sub(start); dur.Seconds() > 3 {
				return
			} else {
				time.Sleep(time.Duration(math.Floor(float64(3)-dur.Seconds())+1) * time.Second)
			}
		}(atomic.LoadInt64(&current))
	}
	swg.Wait()
	return nil
}
