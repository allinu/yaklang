package decompiler

import (
	"slices"

	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/javaclassparser/decompiler/core"
	"github.com/yaklang/yaklang/common/javaclassparser/decompiler/core/statements"
	"github.com/yaklang/yaklang/common/javaclassparser/decompiler/core/values"
	"github.com/yaklang/yaklang/common/javaclassparser/decompiler/rewriter"
	"github.com/yaklang/yaklang/common/utils"
)

func ParseBytesCode(decompiler *core.Decompiler) (res []statements.Statement, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = utils.ErrorStack(e)
		}
	}()
	err = decompiler.ParseSourceCode()
	if err != nil {
		return nil, err
	}
	err = rewriter.CheckNodesIsValid(decompiler.RootNode)
	if err != nil {
		return nil, err
	}

	statementManager := rewriter.NewRootStatementManager(decompiler.RootNode)
	statementManager.SetId(decompiler.CurrentId)
	statementManager.MergeIf()
	allNodes := []*core.Node{}
	core.WalkGraph[*core.Node](decompiler.RootNode, func(node *core.Node) ([]*core.Node, error) {
		allNodes = append(allNodes, node)
		return node.Next, nil
	})
	slices.Reverse(allNodes)
	for _, node := range allNodes {
		if v, ok := node.Statement.(*statements.ConditionStatement); ok {
			if v.Callback != nil {
				v.Callback(v.Condition)
				allNext := slices.Clone(node.Next)
				for _, nextNode := range allNext {
					node.RemoveNext(nextNode)
				}
				for _, sourceNode := range slices.Clone(node.Source) {
					sourceNode.RemoveNext(node)
					for _, n := range allNext {
						sourceNode.AddNext(n)
					}
				}
			}
		}
	}

	err = statementManager.Rewrite()
	if err != nil {
		return nil, err
	}
	nodes, err := statementManager.ToStatements(func(node *core.Node) bool {
		return true
	})
	nodes = funk.Filter(nodes, func(item *core.Node) bool {
		_, ok := item.Statement.(*statements.StackAssignStatement)
		return !ok
	}).([]*core.Node)
	if err != nil {
		return nil, err
	}
	sts := core.NodesToStatements(nodes)
	params := []*values.JavaRef{}
	for _, v := range decompiler.Params {
		if ref, ok := v.(*values.JavaRef); ok {
			params = append(params, ref)
		}
	}
	rewriter.RewriteVar(&sts, decompiler.BodyStartId, params)
	return sts, nil
}
