package main

import (
    "fmt"
    "testing"
    . "github.com/franela/goblin"
    "log"
    //"net/url"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    //"strings"
)

type Person struct {
        FirstName string
        LastName string
        Email string
        Password string
}

func Test(t *testing.T) {

    per := &Person{FirstName:"aaab",LastName:"bbbb",Email:"star9",Password:"star"}
    b, err := json.Marshal(per)
    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post("http://localhost:8080/v1/member/register", "application/json", bytes.NewBuffer(b))

   // resp, err := http.PostForm("http://localhost:8080/v1/member/register",url.Values(b))
    if err != nil{
        log.Fatal(err)
    }
    fmt.Println(resp)

    test1, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    
    check1 := (`{"Success":true,"Desc":"xxx"}`)

    g := Goblin(t)
    g.Describe("Numbers", func() {
        g.It("Should add two numbers ", func() {
            g.Assert(string(test1)).Equal(string(check1))
        })
        // g.It("Should match equal numbers", func() {
        //     g.Assert(2).Equal(4)
        // })
        // g.It("Should substract two numbers")
    })
}