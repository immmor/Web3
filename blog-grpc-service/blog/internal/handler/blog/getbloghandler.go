package blog

import (
	"net/http"

	"blog/blog/internal/logic/blog"
	"blog/blog/internal/svc"
	"blog/blog/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBlogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetBlogReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := blog.NewGetBlogLogic(r.Context(), svcCtx)
		resp, err := l.GetBlog(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
