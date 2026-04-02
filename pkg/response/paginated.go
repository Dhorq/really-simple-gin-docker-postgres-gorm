package response

import "github.com/gin-gonic/gin"

type PaginatedResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Page      int         `json:"page"`
	Limit     int         `json:"limit"`
	Total     int64       `json:"total"`
	TotalPage int         `json:"total_page"`
	Data      interface{} `json:"data"`
}

func SuccessPaginated(c *gin.Context, page, limit int, total int64, data interface{}) {
	totalPage := int(total) / limit
	if int(total)%limit > 0 {
		totalPage++
	}

	c.JSON(200, PaginatedResponse{
		Status:    200,
		Message:   "success",
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: totalPage,
		Data:      data,
	})
}
