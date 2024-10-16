package controler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Update struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var update Update

		if err := ctx.ShouldBindJSON(&update); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if update.Name == "" || update.Email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Name or Eamil connot be empty"})
			return
		}

		query := "UPDATE bdname SET name = ? , email = ? WHERE id = ?"

		_, err = db.Exec(query, update.Name, update.Email, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Filled to update user"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"messag": "User update successfully"})
	}

}
