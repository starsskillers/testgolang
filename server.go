package main
import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/satori/go.uuid"
    "github.com/garyburd/redigo/redis"
   "github.com/gorilla/mux"

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

func handler(writer http.ResponseWriter, request *http.Request){
	

	dec := json.NewDecoder(request.Body)
	var m Person
	err := dec.Decode(&m)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%v: %v: %v: %v\n", m.FirstName, m.LastName, m.Email, m.Password)


	session, err := mgo.Dial("localhost:27017")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    c := session.DB("test").C("Person")
    result := Person{}
    err = c.Find(bson.M{"email": m.Email}).One(&result)
    if err != nil {
        err = c.Insert(&Person{m.FirstName, m.LastName,m.Email, m.Password})
    	if err != nil {
    	        log.Fatal(err)
    	}
    	sen := &Sen{Success:true,Desc:"created User"}
    	b, err := json.Marshal(sen)
    	if err != nil {
        	log.Fatal(err)
        }
    	fmt.Fprintf(writer, "%s", b)
   
    } else {
    	writer.WriteHeader(http.StatusBadRequest)
    	sen := &Sen{Success:false,Desc:"Email used already"}
    	b, err := json.Marshal(sen)
    	if err != nil {
        	log.Fatal(err)
        }
    	fmt.Fprintf(writer, "%s", b)
    }

}

func handler2(writer http.ResponseWriter, request *http.Request){
	dec := json.NewDecoder(request.Body)
	var m Person
	err := dec.Decode(&m)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%v: %v: %v: %v\n", m.FirstName, m.LastName, m.Email, m.Password)
	session, err := mgo.Dial("localhost:27017")
    if err != nil {
        panic(err)
    }
    defer session.Close()

    conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()


    session.SetMode(mgo.Monotonic, true)

    c := session.DB("test").C("Person")
    result := Person{}
    err = c.Find(bson.M{"email": m.Email}).One(&result)
    if err != nil {
    	writer.WriteHeader(http.StatusBadRequest)
        sen := &Sen{Success:false,Desc:"wrong Email"}
    	b, err := json.Marshal(sen)
    	if err != nil {
        	log.Fatal(err)
        }
    	fmt.Fprintf(writer, "%s", b)
    } else {
        if result.Password != m.Password {
            writer.WriteHeader(http.StatusBadRequest)

            sen := &Sen{Success:false,Desc:"wrong Password"}
    		b, err := json.Marshal(sen)
    		if err != nil {
        		log.Fatal(err)
        	}
    		fmt.Fprintf(writer, "%s", b)

        } else {

        	u1 := uuid.NewV4()

			conn.Send("SET",u1.String(), m.Email)
			conn.Send("EXPIRE",u1.String(), 120)
			conn.Flush()
			writer.Header().Add("sessionId",u1.String())

            sen := &Sen{Success:true,Desc:"login successful"}
            b, err := json.Marshal(sen)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Fprintf(writer, "%s : ", b)
        }
    }	
}

func handler3(writer http.ResponseWriter, request *http.Request){

	//fmt.Fprintf(writer, "{%s}", request.Context().Value("sessionId"))
	//request.Context().Value("sessionId")



	vars := mux.Vars(request)
	sess := vars["sessionId"]
	 //fmt.Fprintf(writer, "{%s}", sess)

	//link all database
	session, err := mgo.Dial("localhost:27017")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

	dec := json.NewDecoder(request.Body)
	var m Result
	err = dec.Decode(&m)
	if err != nil {
		panic(err)
	}



   	//fmt.Println(sess)
   	//fmt.Println(m)
    conn.Send("GET",sess)
    conn.Flush()
    email1,_:= conn.Receive()
    d ,_:= redis.String(email1,nil)
    fmt.Println(d)


 	c := session.DB("test").C("Person")
    result := Person{}

    err = c.Find(bson.M{"email": d}).One(&result)
    if err != nil {
        loa := &Load{Success:false,Desc:"change fail worng sessionID",Result1: Result{FirstName:result.FirstName,LastName:result.LastName}}
    	b, err := json.Marshal(loa)
    	if err != nil {
    	    log.Fatal(err)
    	}
    	fmt.Fprintf(writer, "%s", b)      
    } else {
    	//colQuerier := bson.M{"email": result.Email}
		change := bson.M{"$set": bson.M{"firstname": m.FirstName, "lastname": m.LastName}}
		err = c.Update(result, change)
		if err != nil {
			log.Fatal(err)
		}
    	loa := &Load{Success:true,Desc:"change success",Result1: Result{FirstName:m.FirstName,LastName:m.LastName}}
   	 	b, err := json.Marshal(loa)
    	if err != nil {
    	    log.Fatal(err)
    	}
    	fmt.Fprintf(writer, "%s", b)
    }
    //{“success”:true|false,”desc”:”xxx”,result:{“firstName”:”xxx”,”lastName”:”xxx”}}

}


func main(){
	http.HandleFunc("/v1/member/register", handler)
	http.HandleFunc("/v1/member/login", handler2)

	r := mux.NewRouter()
	r.HandleFunc("/v1/member/{sessionId}", handler3)
	http.Handle("/", r)

	// fmt.Println("ListenAndServe :8080")

	http.ListenAndServe(":8080",nil)
}

