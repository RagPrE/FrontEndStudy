package main

import (
	"fmt"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type OwnModel struct {
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-";sql:"index"`
}

type Book struct {
	//gorm.Model
	OwnModel
	//ID     int    `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Isbn   string `json:"isbn"`
	Qty    int    `json:"qty"`
}

/*
type Book struct {
	gorm.Model
	//ID     int
	Name   string
	Author string
	Isbn   string
	Qty    int
}*/

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

var db *gorm.DB
var err error

func main() {
	db2, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("created")
	}
	_, err = db2.Exec("CREATE DATABASE go_test")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created database..")
	}

	db, err = gorm.Open("mysql", "root:root@/go_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.AutoMigrate(&Book{})
	db.Create(&Book{Name: "book1", Author: "tom", Isbn: "abc", Qty: 100})
	db.Create(&Book{Name: "book2", Author: "jim", Isbn: "abc", Qty: 100})
	router := gin.Default()
	router.StaticFS("/static", http.Dir("assets"))
	router.GET("/books", getBook)
	router.GET("/book/:id", getOneBook)
	router.PUT("/book", createBook)
	router.Run(":9025")
}

func getBook(c *gin.Context) {
	//c.JSON(http.StatusOK, m)
	books := make([]Book, 0)
	db.Find(&books)
	c.JSON(http.StatusOK, books)
}

func getOneBook(c *gin.Context) {
	id := c.Param("id")
	var book Book
	if err := db.Where("id = ?", id).First(&book).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, book)
	}
}

func createBook(c *gin.Context) {
	//print(c)
	var book Book
	//book.Name = c.name
	//book.Author = c.author
	c.BindJSON(&book)
	//fmt.Println(book)
	db.Create(&book)
	c.JSON(200, book)
}
