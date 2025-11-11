package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mrbrist/go-blog-aggregator/internal/config"
	"github.com/mrbrist/go-blog-aggregator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func main() {
	// GET CONFIG FILE
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// DATABASE CONNECTION
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	dbQueries := database.New(db)

	// SET STATE AND HANDLE COMMANDS
	s := state{cfg: cfg, db: dbQueries}
	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}
	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	// REGISTER COMMANDS
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)

	// -----------------

	cmd := command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	// RUN COMMAND
	if err := cmds.run(&s, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	name := cmd.Args[0]
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User does not exist!")
		return err
	}

	err1 := config.SetUser(s.cfg, user.Name)
	if err1 != nil {
		return err1
	}
	fmt.Println("Username has been set!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no name given")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.Args[0]})
	if err != nil {
		return err
	}
	err1 := config.SetUser(s.cfg, user.Name)
	if err1 != nil {
		return err1
	}
	fmt.Println("User was created!")
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database has been reset!")
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
}
