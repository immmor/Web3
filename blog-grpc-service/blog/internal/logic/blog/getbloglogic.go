package blog

import (
	"context"

	"blog/blog/internal/svc"
	"blog/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlogLogic {
	return &GetBlogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlogLogic) GetBlog(req *types.GetBlogReq) (resp *types.GetBlogResp, err error) {
	// todo: add your logic here and delete this line
	
	// 临时返回一个示例blog对象
	return &types.GetBlogResp{
		Blog: types.Blog{
			Id:        req.Id,
			Title:     "示例博客标题",
			Content:   "这是示例博客内容",
			AuthorId:  1,
			CreatedAt: "2024-01-01 10:00:00",
		},
	}, nil
}
