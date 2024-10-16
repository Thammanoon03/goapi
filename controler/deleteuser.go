package controler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"eroor": "Invalid Id"})
			return
		}

		query := "DELETE FROM bdname WHERE id = ?"

		_, err = db.Exec(query, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to database to user"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"error": "user delete successfully"})
	}
}
