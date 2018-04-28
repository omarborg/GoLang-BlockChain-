package main

import (

        "crypto/sha256"
        "encoding/hex"
        "encoding/json"
        "io"
        "log"
        "net/http"
        "os"
        "time"

        "github.com/davecgh/go-spew/spew"
        "github.com/gorilla/mux"
        "github.com/joho/godotenv"
)

//Define struct of each block

type Block struct {

        Index        int      //position of record in blockchain
        TimeStamp    string   //Automatically determined by the time the data is written
        BPM          int      //Beats per minute
        Hash         string   //SHA256 identifier
        PrevHash     string   //SHA256 of previous record in chain
}

var Blockchain []Block
type message struct {

        BPM int
}
//function main
func main() {

    err := godotenv.Load()
    if err != nil {

            log.Fatal(err)
    }

    go func ()  {

             t := time.Now()
             genBlock := Block{0,t.String(), 0, "", ""}
             spew.Dump(genBlock)
             Blockchain = append(Blockchain, genBlock)
    } ()
    log.Fatal(run())
}

//function to run a web server to view the blockchain using Gorilla/mux package
func run() error {

    mux := makeMuxRouter()
    httpAddr := os.Getenv("ADDR")       // Getenv() from the .env file created in the path should have one line defining ADDR with the desired port
    log.Println("Listening on", httpAddr)
    s := &http.Server {

          Addr:             ":" + httpAddr,
          Handler:          mux,
          ReadTimeout:      10 * time.Second,
          WriteTimeout:     10 * time.Second,
          MaxHeaderBytes:   1 << 20,
    }

    if err := s.ListenAndServe(); err != nil {

          return err
    }

    return nil
}

//function to define handlers
func makeMuxRouter() http.Handler {

      muxRouter := mux.NewRouter()
      muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
      muxRouter.HandleFunc("/", handleWriteBlockchain).Methods("POST")
      return muxRouter

}

//function to handle GET requests
func handleGetBlockchain (w http.ResponseWriter, r *http.Request) {

      bytes, err := json.MarshalIndent(Blockchain, "", " ")
      if err != nil {

            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
      }

      io.WriteString(w, string(bytes))
}

//Message struct for POST request
// function to handled POST errors using JSON
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int , payload interface{}) {

      response , err := json.MarshalIndent(payload, "", " ")
      if err != nil {

            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte("HTTP 500: Internal Server Error"))
            return

      }
      w.WriteHeader(code)
      w.Write(response)
}
//function to handle POST request and writeBlock
func handleWriteBlockchain(w http.ResponseWriter, r *http.Request) {

      var m message

      decoder := json.NewDecoder(r.Body)
      if err := decoder.Decode(&m); err != nil {

              respondWithJSON(w,r, http.StatusBadRequest, r.Body)
      }

      defer r.Body.Close()

      newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
      if err != nil {

              respondWithJSON(w,r, http.StatusInternalServerError, m)
              return
      }

      if blockIsValid(newBlock, Blockchain[len(Blockchain)-1]) {

            newBlockchain := append(Blockchain, newBlock)
            chooseChain(newBlockchain)
            spew.Dump(Blockchain)
      }

      respondWithJSON(w, r, http.StatusCreated, newBlock)

}


//function to check validity of blocks
func blockIsValid( newBlock, oldBlock Block) bool {

      if oldBlock.Index + 1 != newBlock.Index {

          return false
      }

      if oldBlock.Hash != newBlock.PrevHash {

          return false
      }

      if calculateHash(newBlock) != newBlock.Hash {

          return false
      }

      return true
}


//function to create the SHA256 Hash
func calculateHash (block Block) string {

      //record represents the block, h generates a new sha256 and writes it to the record & finally hashed returns the value in an encoded string
      record := string(block.Index) + block.TimeStamp + string(block.BPM) + block.PrevHash
      h := sha256.New()
      h.Write([]byte(record))
      hashed := h.Sum(nil)
      return hex.EncodeToString(hashed)

}

//function to generate a block
func generateBlock (oldBlock Block, BPM int) (Block, error) {

      var newBlock Block
      t := time.Now()

      newBlock.Index = oldBlock.Index + 1
      newBlock.TimeStamp = t.String()
      newBlock.BPM = BPM
      newBlock.PrevHash = oldBlock.Hash
      newBlock.Hash = calculateHash(newBlock)

      return newBlock, nil
}

//function to check the length of chain used to determine the most up to date block in case of two blocks are added in the same time
func chooseChain(newBlocks []Block) {

      if len(newBlocks) > len(Blockchain) {

            Blockchain = newBlocks

    }
}
