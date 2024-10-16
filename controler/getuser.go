package controler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Users struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var users []Users

		query := "SELECT id, name, email FROM bdname"
		rows, err := db.Query(query)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var user Users
			if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user data"})
				return
			}
			users = append(users, user)
		}

		if err := rows.Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error"})
			return
		}
		ctx.JSON(http.StatusOK, users)
	}

}
