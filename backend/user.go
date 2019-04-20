package main

import (
      elastic "gopkg.in/olivere/elastic.v3"

      "encoding/json"
      "fmt"
      "net/http"
      "reflect"
      "regexp"
      "time"
      //"io/ioutil"
      //"log"

      "github.com/dgrijalva/jwt-go"
)

const (
      TYPE_USER = "user"
)

var (
      usernamePattern = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
)

type User struct {
      Username string `json:"username"`
      Password string `json:"password"`
      Age int `json:"age"`
      Gender string `json:"gender"`
}

//check User check whether user is valid
func checkUser(username, password string) bool {
     es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
     if err != nil {
     	 fmt.Printf("ES is not setup %v\n", err)
	     panic(err)
       return false
     }

     //Search with a term query
     termQuery := elastic.NewTermQuery("username", username)
     queryResult, err := es_client.Search().
     		  Index(INDEX).
		  Query(termQuery).
		  Pretty(true).
		  Do()
     if err != nil {
     	fmt.Printf("ES query failed %v\n", err)
	     panic(err)
       return false
     }

     var tyu User
     for _, item := range queryResult.Each(reflect.TypeOf(tyu)) {
     	 u:= item.(User)
	 return u.Password == password && u.Username == username 
     }

     return false
}

//add a new user return true if successful
func addUser(user User) bool {
     es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
     if err != nil {
     	fmt.Printf("ES is not setup %v\n", err)
	return false
     }

     termQuery := elastic.NewTermQuery("username", user.Username)
     queryResult, err := es_client.Search().
     		  Index(INDEX).
		  Query(termQuery).
		  Pretty(true).
		  Do()
     if err != nil {
     	fmt.Printf("ES query failed %v\n", err)
        return false
     }

     if queryResult.TotalHits()>0 {
     	fmt.Printf("User %s already exists, cannot dupilcate user. \n", user.Username)
	return false
     }

     _, err = es_client.Index().
     	Index(INDEX).
	Type(TYPE_USER).
	Id(user.Username).
	BodyJson(user).
	Refresh(true).
	Do()
     if err != nil {
     	fmt.Printf("ES save user failed %v\n", err)
	return false
     }

     return true
}		

// if signup is successful, a new session is created
func signupHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received one signup request")
    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

    // body, err := ioutil.ReadAll(r.Body)
    // if err != nil {
    //     panic(err)
    // }


    // var u User
    // err = json.Unmarshal(body, &u)
    // if err != nil {
    //   panic(err)
    //   return
    // }

      decoder := json.NewDecoder(r.Body)
      var u User
      if err := decoder.Decode(&u); err != nil {
             panic(err)
             return
      }

    if u.Username != "" && u.Password != "" && usernamePattern(u.Username) {
      if addUser(u) {
        fmt.Println("User added successfully.")
        w.Write([]byte("User added successfully"))
      } else {
        fmt.Println("failed to add a new User.")
        http.Error(w, "failed to add a new User", http.StatusInternalServerError)
      }
    } else {
      fmt.Println("Empty password or username or invalid username.")
      http.Error(w, "Empty password or username or invalid username.", http.StatusInternalServerError)
    }



     
} 



// If login is successful, a new token is created.
func loginHandler(w http.ResponseWriter, r *http.Request) {
      fmt.Println("Received one login request")
      w.Header().Set("Content-Type", "text/plain")
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

      // body, err := ioutil.ReadAll(r.Body)
      // if err != nil {
      //   panic(err)
      //   return
      // }

      // var u User
      // err = json.Unmarshal(body, &u)
      // if err != nil {
      //   panic(err)
      //   return
      // }





      decoder := json.NewDecoder(r.Body)
      var u User
      if err := decoder.Decode(&u); err != nil {
             panic(err)
             return
      }

      if checkUser(u.Username, u.Password) {
             token := jwt.New(jwt.SigningMethodHS256)
             claims := token.Claims.(jwt.MapClaims)
             /* Set token claims */
             claims["username"] = u.Username
             claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

             /* Sign the token with our secret */
             tokenString, _ := token.SignedString(mySigningKey)

             /* Finally, write the token to the browser window */
             w.Write([]byte(tokenString))
      } else {
             fmt.Println("Invalid password or username.")
             http.Error(w, "Invalid password or username", http.StatusForbidden)
      }



}

