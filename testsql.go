package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "time"
)

type History struct {
  gorm.Model
  FirstName string
  LastName string
  Email string
  TimeStamp time.Time
}

func main() {
  db, err := gorm.Open("mysql","root:root@(localhost:3306)/History?charset=utf8&parseTime=True&loc=Local")
  if err != nil {
    panic("failed to connect database")
  }
  defer db.Close()

  // Migrate the schema
  db.AutoMigrate(&History{})

  // Create
  db.Create(&History{FirstName: "ffff", LastName: "lllll", Email:"eeee", TimeStamp: time.Now()})

  // Read
  // var his History
  // db.First(&his, 1) // find product with id 1
  // db.First(&his, "code = ?", "L1212") // find product with code l1212

  // // Update - update product's price to 2000
  // db.Model(&product).Update("Price", 2000)

  // Delete - delete product
  //db.Delete(&product)
}