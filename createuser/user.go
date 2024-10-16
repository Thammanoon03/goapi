package createuser

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newUser User

		if err := ctx.ShouldBindJSON(&newUser); err != nil {
			log.Println("Failed to bind JSON:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := "INSERT INTO bdname (name,email) VALUES (?,?)"

		_, err := db.Exec(query, newUser.Name, newUser.Email)
		if err != nil {
			log.Println("Failed to insert user:", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "User inserted successfully"})
		log.Println("User inserted successfully")
	}
}
