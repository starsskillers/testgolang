package main
import (
	"fmt"
	"net/http"
	"encoding/json"
	// "log"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/satori/go.uuid"
    "github.com/garyburd/redigo/redis"
   	"github.com/gorilla/mux"
    // "github.com/jinzhu/gorm"
    // _ "github.com/jinzhu/gorm/dialects/mysql"
    // "time"
)

type Person struct {
    FirstName string
    LastName string
    Email string
    Password string
}
type Sen struct {
    Success bool
    Desc string
}
type Result struct{
	FirstName string
    LastName string
}
type Load struct {
    Success bool
    Desc string
    Result1 Result 
}
// type History struct {
//   	gorm.Model
//   	FirstName string
//   	LastName string
//   	Email string
//   	TimeStamp time.Time
// }

var mongo *mgo.Session
var connRedis redis.Conn

func connectMongoDB () {
	var err error
	mongo, err = mgo.Dial("localhost:27017")
    if err != nil {
        fmt.Println("can't connect mongoDB")
    }
    mongo.SetMode(mgo.Monotonic, true)
}
func connectRedis () {
	var err error
	connRedis, err = redis.Dial("tcp", "localhost:6379")
    if err != nil {
        fmt.Println("can't connect redis")
    }
}

func handler(writer http.ResponseWriter, request *http.Request){
	

	decode := json.NewDecoder(request.Body)
	var jsonInput Person
	err := decode.Decode(&jsonInput)
	if err != nil {
		fmt.Printf("error can't decode JSON")
	}
   	category := mongo.DB("test").C("Person")
    result := Person{}
    err = category.Find(bson.M{"email": jsonInput.Email}).One(&result)
    if err != nil {
        err = category.Insert(&Person{jsonInput.FirstName, jsonInput.LastName,jsonInput.Email, jsonInput.Password})
    	if err != nil {
    	        fmt.Printf("error can't insert data")
    	}
    	sen := &Sen{Success:true,Desc:"created User"}
    	jsonOutput, err := json.Marshal(sen)
    	if err != nil {
        	fmt.Printf("error can't sent JSON in true case")
        }
    	fmt.Fprintf(writer, "%s", jsonOutput)
   
    } else {
    	writer.WriteHeader(http.StatusBadRequest)
    	sen := &Sen{Success:false,Desc:"Email used already"}
    	jsonOutput, err := json.Marshal(sen)
    	if err != nil {
        	fmt.Printf("error 3")
        }
    	fmt.Fprintf(writer, "%s", jsonOutput)
    }

}

func handler2(writer http.ResponseWriter, request *http.Request){
	decode := json.NewDecoder(request.Body)
	var jsonInput Person
	err := decode.Decode(&jsonInput)
	if err != nil {
		fmt.Printf("error can't decoed JSON")
	}

	connectRedis()

    category := mongo.DB("test").C("Person")
    result := Person{}
    err = category.Find(bson.M{"email": jsonInput.Email}).One(&result)
    if err != nil {
    	writer.WriteHeader(http.StatusBadRequest)
        sen := &Sen{Success:false,Desc:"wrong Email"}
    	jsonOutput, err := json.Marshal(sen)
    	if err != nil {
        	fmt.Printf("error can't sent JSON wrong email")
        }
    	fmt.Fprintf(writer, "%s", jsonOutput)
    } else {
        if result.Password != jsonInput.Password {
            writer.WriteHeader(http.StatusBadRequest)

            sen := &Sen{Success:false,Desc:"wrong Password"}
    		jsonOutput, err := json.Marshal(sen)
    		if err != nil {
        		fmt.Printf("error can't sent JSON wrong password")
        	}
    		fmt.Fprintf(writer, "%s", jsonOutput)

        } else {

        	u1 := uuid.NewV4()

			connRedis.Send("SET",u1.String(), jsonInput.Email)
			connRedis.Send("EXPIRE",u1.String(), 120)
			connRedis.Flush()
			writer.Header().Add("sessionId",u1.String())

            sen := &Sen{Success:true,Desc:"login successful"}
            jsonOutput, err := json.Marshal(sen)
            if err != nil {
                fmt.Printf("error can't sent JSON login successful")
            }
            fmt.Fprintf(writer, "%s : ", jsonOutput)
        }
    }	
}

func handler3(writer http.ResponseWriter, request *http.Request){

	vars := mux.Vars(request)
	sess := vars["sessionId"]

	connectRedis()

   // 	db, err := gorm.Open("mysql","root:root@(localhost:3306)/History?charset=utf8&parseTime=True&loc=Local")
  	// if err != nil {
   //  	panic("failed to connect database")
  	// }
  	// defer db.Close()


	decode := json.NewDecoder(request.Body)
	var jsonInput Result
	err := decode.Decode(&jsonInput)
	if err != nil {
		fmt.Printf("error can't decode JSON")
	}

    connRedis.Send("GET",sess)
    connRedis.Flush()
    email1,_:= connRedis.Receive()
    redisdata ,_:= redis.String(email1,nil)

 	category := mongo.DB("test").C("Person")
    result := Person{}

    err = category.Find(bson.M{"email": redisdata}).One(&result)
    if err != nil {
        loa := &Load{Success:false,Desc:"change fail worng sessionID",Result1: Result{FirstName:result.FirstName,LastName:result.LastName}}
    	jsonOutput, err := json.Marshal(loa)
    	if err != nil {
    	    fmt.Printf("error can't sent JSON wrong sessionID")
    	}
    	fmt.Fprintf(writer, "%s", jsonOutput)      
    } else {
		change := bson.M{"$set": bson.M{"firstname": jsonInput.FirstName, "lastname": jsonInput.LastName}}
		err = category.Update(result, change)
		if err != nil {
			fmt.Printf("error can't update category MongoDB")
		}
    	loa := &Load{Success:true,Desc:"change success",Result1: Result{FirstName:jsonInput.FirstName,LastName:jsonInput.LastName}}
   	 	jsonOutput, err := json.Marshal(loa)
    	if err != nil {
    	    fmt.Printf("error can't sent JSON success change name")
    	}
    	fmt.Fprintf(writer, "%s", jsonOutput)
    }
    //mysql code

  //   // Migrate(โยกย้าย) the schema 
 	// db.AutoMigrate(&History{})

  // 	// Create
  // 	db.Create(&History{FirstName: result.FirstName, LastName: result.LastName, Email:result.Email, TimeStamp: time.Now()})


}


func main(){
	connectMongoDB()
	http.HandleFunc("/v1/member/register", handler)
	http.HandleFunc("/v1/member/login", handler2)
	r := mux.NewRouter()
	r.HandleFunc("/v1/member/{sessionId}", handler3)
	http.Handle("/", r)
	http.ListenAndServe(":8080",nil)
}

