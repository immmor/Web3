package blog

import (
	"context"

	"blog/blog/internal/svc"
	"blog/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBlogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBlogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBlogsLogic {
	return &ListBlogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBlogsLogic) ListBlogs(req *types.ListBlogsReq) (resp *types.ListBlogsResp, err error) {
	// todo: add your logic here and delete this line
	
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	
	// 临时返回示例数据
	return &types.ListBlogsResp{
		Blogs: []types.Blog{
			{
				Id:        1,
				Title:     "示例博客1",
				Content:   "这是第一个示例博客的内容",
				AuthorId:  1,
				CreatedAt: "2024-01-01 10:00:00",
			},
			{
				Id:        2,
				Title:     "示例博客2", 
				Content:   "这是第二个示例博客的内容",
				AuthorId:  1,
				CreatedAt: "2024-01-01 11:00:00",
			},
		},
		Total: 2,
	}, nil
}
