package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	weather "myproject/contr"
	"myproject/controler"
	"myproject/createuser"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
)

type PriceResponse struct {
	Response Response `json:"response"`
}

type Response struct {
	Date       string `json:"date"`
	UpdateTime string `json:"update_time"`
	Price      Price  `json:"price"`
}

type Price struct {
	GoldBar Value  `json:"gold_bar"`
	Change  Change `json:"change"`
}

type Value struct {
	Buy  string `json:"buy"`
	Sell string `json:"sell"`
}

type Change struct {
	ComparePrevious  string `json:"compare_previous"` // Fixed typo here
	CompareYesterday string `json:"compare_yesterday"`
}

func getGoldPrice(c *gin.Context) {

	tokenStr := c.GetHeader("Authorization")
	token, err := jwt.ParseWithClaims(tokenStr, &claim{}, func(token *jwt.Token) (interface{}, error) {
		return jwtkey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	claims, ok := token.Claims.(*claim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}

	url := "https://api.chnwt.dev/thai-gold-api/latest"
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch gold price data"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch gold price data"})
		return
	}

	var priceResponse PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding gold data"})
		return
	}

	if claims.Role == "user" {
		c.JSON(http.StatusOK, gin.H{
			"data":          priceResponse.Response.Date,
			"update_time":   priceResponse.Response.UpdateTime,
			"gold_bar_sell": priceResponse.Response.Price.GoldBar.Sell,
		})
	} else if claims.Role == "admin" {
		c.JSON(http.StatusOK, priceResponse)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permission"})
	}
}

var (
	db     *sql.DB
	jwtkey = []byte("my_secret_key")
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type claim struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func login(ctx *gin.Context) {
	var creds Credentials
	if err := ctx.ShouldBindJSON(&creds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var passwordInDB, roleInDB string
	err := db.QueryRow("SELECT password , role FROM login WHERE username = ?", creds.Username).Scan(&passwordInDB, &roleInDB)
	if err == sql.ErrNoRows || passwordInDB != creds.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &claim{
		Username: creds.Username,
		Role:     roleInDB,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.GetHeader("Authorization")
		if tokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			ctx.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &claim{}, func(token *jwt.Token) (interface{}, error) {
			return jwtkey, nil
		})
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func roleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.GetHeader("Authorization")
		if tokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			ctx.Abort()
			return
		}
		token, err := jwt.ParseWithClaims(tokenStr, &claim{}, func(t *jwt.Token) (interface{}, error) {
			return jwtkey, nil
		})
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}
		claims, ok := token.Claims.(*claim)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			ctx.Abort()
			return
		}
		for _, role := range allowedRoles {
			if claims.Role == role {
				ctx.Next()
				return
			}
		}
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you do not have the rwquired rple"})
		ctx.Abort()
	}
}

func main() {
	dsn := "root@tcp(127.0.0.1:3306)/node_crud_db"
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MySQL successfully")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},  // อนุญาตแหล่งที่มา (origin)
		AllowMethods:     []string{"GET", "POST", "OPTIONS"}, // อนุญาตวิธีการ HTTP
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/api/login", login)
	r.GET("/api/weather", weather.GetWeather)
	r.GET("/api/gold-price", getGoldPrice)
	r.GET("/api/getuser", controler.GetUserHandler(db))
	r.POST("/api/createuser", createuser.CreateUser(db))
	r.PUT("/api/updateuser/:id", controler.UpdateUser(db))
	r.DELETE("api/deleteuser/:id", controler.DeleteUser(db))

	protected := r.Group("/api")
	protected.Use(authMiddleware())

	//roleMiddleware เพื่อให้เฉพาะผู้ใช้ที่มี role เป็น "admin" สามารถเข้าถึงได้
	protected.GET("/admin", roleMiddleware("admin"), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Welcome, admin!"})
	})

	// สำหรับ role user
	protected.GET("/user", roleMiddleware("user"), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Welcome, user!"})
	})

	fmt.Println("Before running the server")
	r.Run(":5000")
	fmt.Println("After running the server")
}
