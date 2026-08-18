package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	ID       int64  `json:"id"`
	AuthorID int64  `json:"author_id"`
	ChatID   int64  `json:"chat_id"`
	Text     string `json:"text"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func userByName(name string) (User, error) {
	var user User
	row := db.QueryRow("SELECT * FROM users WHERE name = ?", name)
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return user, err
		}
		return user, fmt.Errorf("userByName %q: %v", name, err)
	}
	return user, nil
}

func checkPassword(usr_id int64, pass string) (bool, error) {
	hash := sha256.Sum256([]byte(pass))

	correct := make([]byte, 32)
	row := db.QueryRow("SELECT pass_hash FROM users_pass WHERE id = ?", usr_id)
	if err := row.Scan(&correct); err != nil {
		return false, fmt.Errorf("checkPassword: %v", err)
	}
	return bytes.Equal(hash[:], correct), nil
}

func setPassword(usr_id int64, pass string) error {
	hash := sha256.Sum256([]byte(pass))
	fmt.Printf("%+v\n", hash)
	_, err := db.Exec("INSERT INTO users_pass (id, pass_hash) VALUES (?, ?)",
		usr_id, hash[:])
	if err != nil {
		return fmt.Errorf("setPassword %d: %v", usr_id, err)
	}
	fmt.Printf("Done\n")
	return nil
}

func userById(usr_id int64) (User, error) {
	var user User
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", usr_id)
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return user, err
		}
		return user, fmt.Errorf("userById %d: %v", usr_id, err)
	}
	return user, nil
}

func allUsers() ([]User, error) {
	var users []User
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("allUsers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var usr User
		if err := rows.Scan(&usr.ID, &usr.Name); err != nil {
			return nil, fmt.Errorf("allUsers: %v", err)
		}
		users = append(users, usr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("allUsers: %v", err)
	}
	return users, nil
}

func addUser(name string) (int64, error) {
	result, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return 0, fmt.Errorf("addUser %s: %v", name, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addUser %s: %v", name, err)
	}
	return id, nil
}

func chatByUserId(usr_id int64) ([]Chat, error) {
	var chats []Chat
	rows, err := db.Query("SELECT id FROM chats WHERE member_id = ?", usr_id)
	if err != nil {
		return nil, fmt.Errorf("chatByUserId %d: %v", usr_id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID); err != nil {
			return nil, fmt.Errorf("chatByUserId %d: %v", usr_id, err)
		}

		row := db.QueryRow("SELECT name FROM chat_names WHERE id = ?", chat.ID)
		if err := row.Scan(&chat.Name); err != nil {
			return nil, fmt.Errorf("chatByUserId %d: %v", usr_id, err)
		}
		chats = append(chats, chat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chatByUserId %d: %v", usr_id, err)
	}
	return chats, nil
}

func chatHistoryById(chat_id int64) ([]Message, error) {
	var history []Message
	rows, err := db.Query("SELECT * FROM messages WHERE chat_id = ?", chat_id)
	if err != nil {
		return nil, fmt.Errorf("chatHistoryById %d: %v", chat_id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.AuthorID, &msg.ChatID, &msg.Text); err != nil {
			return nil, fmt.Errorf("chatHistoryById %d: %v", chat_id, err)
		}
		history = append(history, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chatHistoryById %d: %v", chat_id, err)
	}
	return history, nil
}

func chatMembersById(chat_id int64) ([]User, error) {
	var members []User
	rows, err := db.Query("SELECT member_id FROM chats WHERE id = ?", chat_id)
	if err != nil {
		return nil, fmt.Errorf("chatMembersById %d: %v", chat_id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var usr_id int64
		if err := rows.Scan(&usr_id); err != nil {
			return nil, fmt.Errorf("chatMembersById %d: %v", chat_id, err)
		}

		usr, err := userById(usr_id)
		if err != nil {
			return nil, fmt.Errorf("chatMembersById %d: %v", chat_id, err)
		}
		members = append(members, usr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chatMembersById %d: %v", chat_id, err)
	}
	return members, nil
}

func sendMesage(msg Message) error {
	_, err := db.Exec("INSERT INTO messages (author_id, chat_id, text) VALUES (?, ?, ?)",
		msg.AuthorID, msg.ChatID, msg.Text)	
	if err != nil {
		return fmt.Errorf("sendMesage %v", err)
	}
	return nil
}

func makeChat(chatName string, usersIds []int64) error {
	result, err := db.Exec("INSERT INTO chat_names (name) VALUES (?)", chatName)
	if err != nil {
		return fmt.Errorf("makeChat %s: %v", chatName, err)
	}
	chatId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("makeChat %s: %v", chatName, err)
	}
	for _, usrId := range usersIds {
		_, err = db.Exec("INSERT INTO chats (id, member_id) VALUES (?, ?)", chatId, usrId)
		if err != nil {
		  return fmt.Errorf("makeChat %s: %v", chatName, err)
		}
	}
	return nil
}

func main() {
	// DB
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbDb := os.Getenv("MYSQL_DATABASE")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbUser, dbPassword, dbHost,
		dbPort, dbDb)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error when connecting to db: %v", err)
	}
	defer db.Close()
	fmt.Println("Connected to db")

	// http
	http.HandleFunc("/", handler)
	http.HandleFunc("/healthy", healthyHandler)

	http.HandleFunc("/api/v1/chats", handlerChats)
	http.HandleFunc("/api/v1/users", handlerUsers)
	http.HandleFunc("/api/v1/registration", handlerRegistration)
	http.HandleFunc("/api/v1/login", handlerLogin)
	http.HandleFunc("/api/v1/chat_history", handlerChatHistory)
	http.HandleFunc("/api/v1/chat_members", handlerChatMembers)
	http.HandleFunc("/api/v1/make_chat", handlerMakeChat)
	http.HandleFunc("/api/v1/send_msg", handlerSendMessage)

	fmt.Println("Server listening on 8080")
	http.ListenAndServe(":8080", nil)
}

// handlers
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Url: %s\n", r.URL.String())
	fmt.Fprintf(w, "Wrong Way")
}
func healthyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func handlerChats(w http.ResponseWriter, r *http.Request) {
	var (
		query       = r.URL.Query()
		user_id_str = query.Get("userid")
	)
	user_id, err := strconv.ParseInt(user_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	fmt.Printf("Request from %d\n", user_id)
	chats_lst, err := chatByUserId(user_id)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	json_chats_lst, err := json.Marshal(chats_lst)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true, \"chats\": %s}", string(json_chats_lst))
}

func handlerUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request users list\n")
	usr_lst, err := allUsers()
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	json_lst, err := json.Marshal(usr_lst)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true, \"users\": %s}", string(json_lst))
}

func handlerRegistration(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Attempt to registrate %s\n", r.URL.String())
	var (
		query    = r.URL.Query()
		new_user = query.Get("newname")
		passw = query.Get("passw")
	)
	_, err := userByName(new_user)
	if err == nil {
	}
	if err != sql.ErrNoRows {
		if err != nil {
			fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		} else {
			fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"Existed\"}")
		}
	} else {
		id, err := addUser(new_user)
		if err != nil {
			fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		} else {
			err = setPassword(id, passw)
			if err != nil {
				fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
				return
			}
			fmt.Printf("Send result to front\n")
			fmt.Fprintf(w, "{\"ok\": true, \"id\": %d, \"name\": \"%s\"}", id, new_user)
		}
	}
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	var (
		query = r.URL.Query()
		user_id_str = query.Get("userid")
		pass = query.Get("passw")
	)

	user_id, err := strconv.ParseInt(user_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	check, err := checkPassword(user_id, pass)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
	}
	fmt.Fprintf(w, "{\"ok\": true, \"passed\": %t}", check)
}

func handlerChatHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request history\n")
	var (
		query       = r.URL.Query()
		chat_id_str = query.Get("chatid")
	)
	chat_id, err := strconv.ParseInt(chat_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	history, err := chatHistoryById(chat_id)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	json_history, err := json.Marshal(history)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true, \"history\": %s}", string(json_history))
}

func handlerChatMembers(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request members\n")
	var (
		query       = r.URL.Query()
		chat_id_str = query.Get("chatid")
	)
	chat_id, err := strconv.ParseInt(chat_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	members, err := chatMembersById(chat_id)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	json_members, err := json.Marshal(members)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true, \"members\": %s}", string(json_members))
}

func handlerSendMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Send msg\n")
	var (
		query = r.URL.Query()
		chat_id_str = query.Get("chatid")
		author_id_str = query.Get("authorid")
		text = query.Get("text")
	)
	chat_id, err := strconv.ParseInt(chat_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	author_id, err := strconv.ParseInt(author_id_str, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}

	msg := Message{AuthorID: author_id,
								 ChatID: chat_id,
						 		 Text: text}
	fmt.Printf("Msg: %s\n", msg)
	err = sendMesage(msg)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true}")
}

func handlerMakeChat(w http.ResponseWriter, r *http.Request) {
	var (
		query = r.URL.Query()
		chatName = query.Get("chatname")
		members_id_str = query["membersid"]
	)
	fmt.Printf("makeChat %s with %s\n", chatName, members_id_str)
	members_id := make([]int64, len(members_id_str))
	for i, str_id := range members_id_str {
		id, err := strconv.ParseInt(str_id, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
			return
		}
		members_id[i] = id
	}

	err := makeChat(chatName, members_id)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"msg\": \"%v\"}", err)
		return
	}
	fmt.Fprintf(w, "{\"ok\": true}")
}

