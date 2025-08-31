package blog

import (
	"net/http"

	"blog/blog/internal/logic/blog"
	"blog/blog/internal/svc"
	"blog/blog/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListBlogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListBlogsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := blog.NewListBlogsLogic(r.Context(), svcCtx)
		resp, err := l.ListBlogs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
