package order

import (
	"errors"
	"five/internal/logic/order"
	"five/internal/svc"
	"five/internal/types"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Order
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		err := l.CreateOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")
		reason := r.URL.Query().Get("reason")
		if orderID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("order_id is required"))
			return
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		err := l.CancelOrder(orderID, reason)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

func FillOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")
		fillPrice, _ := strconv.ParseFloat(r.URL.Query().Get("price"), 64)
		fillAmount, _ := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)
		feeRate, _ := strconv.ParseFloat(r.URL.Query().Get("fee_rate"), 64)

		if orderID == "" || fillPrice <= 0 || fillAmount <= 0 {
			httpx.ErrorCtx(r.Context(), w, errors.New("invalid parameters"))
			return
		}

		// 默认手续费率 0.1%
		if feeRate <= 0 {
			feeRate = 0.001
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		err := l.FillOrder(orderID, fillPrice, fillAmount, feeRate)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

func GetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")
		if orderID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("order_id is required"))
			return
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		result, err := l.GetOrder(orderID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJson(w, result)
		}
	}
}

func GetUserOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
		status := types.OrderStatus(r.URL.Query().Get("status"))

		if userID <= 0 {
			httpx.ErrorCtx(r.Context(), w, errors.New("user_id is required"))
			return
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		result, err := l.GetUserOrders(userID, status)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJson(w, result)
		}
	}
}

func GetOrderTradesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("order_id")
		if orderID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("order_id is required"))
			return
		}

		l := order.NewOrderLogic(r.Context(), svcCtx)
		result, err := l.GetOrderTrades(orderID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJson(w, result)
		}
	}
}
