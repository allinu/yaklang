package yakgrpc

import (
	"context"

	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

func (s *Server) QuerySSAPrograms(ctx context.Context, req *ypb.QuerySSAProgramRequest) (*ypb.QuerySSAProgramResponse, error) {
	pagine, programs, err := yakit.QuerySSAProgram(s.GetSSADatabase(), req)
	if err != nil {
		return nil, err
	}
	return &ypb.QuerySSAProgramResponse{
		Pagination: &ypb.Paging{
			Page:  int64(pagine.Page),
			Limit: int64(pagine.Limit),
		},
		Data:  programs,
		Total: int64(pagine.TotalRecord),
	}, nil
}

func (s *Server) UpdateSSAProgram(ctx context.Context, req *ypb.UpdateSSAProgramRequest) (*ypb.DbOperateMessage, error) {
	count, err := yakit.UpdateSSAProgram(s.GetSSADatabase(), req.GetProgramInput())
	return &ypb.DbOperateMessage{
		TableName:    "ssa_programs",
		Operation:    "update",
		EffectRows:   count,
		ExtraMessage: "",
	}, err
}

func (s *Server) DeleteSSAPrograms(ctx context.Context, req *ypb.DeleteSSAProgramRequest) (*ypb.DbOperateMessage, error) {
	var count int
	var err error
	if req.DeleteAll {
		count, err = yakit.DeleteSSAProgram(s.GetSSADatabase(), nil)
	} else if req.GetFilter() != nil {
		count, err = yakit.DeleteSSAProgram(s.GetSSADatabase(), req.GetFilter())
	}
	if err != nil {
		return nil, err
	}
	return &ypb.DbOperateMessage{
		TableName:    "ssa_programs",
		Operation:    "delete",
		EffectRows:   int64(count),
		ExtraMessage: "",
	}, nil
}
