package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"cloud.google.com/go/firestore"
	pb "github.com/odmishien/grpctodo/todo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port      = ":50051"
	projectid = "test"
)

type server struct {
	pb.UnimplementedTodoServiceServer
}

func (s *server) AddTodo(ctx context.Context, in *pb.AddTodoParams) (*pb.TodoObject, error) {
	task := in.GetTask()
	if task == "" {
		return nil, errors.New("invalid argument")
	}
	fs, err := firestore.NewClient(ctx, projectid)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()
	t, _, err := fs.Collection("todos").Add(ctx, map[string]interface{}{
		"task": task,
	})
	return &pb.TodoObject{Id: t.ID, Task: task}, nil
}

func (s *server) RemoveTodo(ctx context.Context, in *pb.RemoveTodoParams) (*pb.RemoveResponse, error) {
	id := in.GetId()
	if id == "" {
		return nil, errors.New("invalid argument")
	}
	fs, err := firestore.NewClient(ctx, projectid)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()
	_, err = fs.Collection("todos").Doc(id).Delete(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return &pb.RemoveResponse{Message: "OK"}, nil
}

func (s *server) GetTodos(ctx context.Context, in *pb.GetTodoParams) (*pb.TodoResponse, error) {
	fs, err := firestore.NewClient(ctx, projectid)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Printf("%#v\n", fs.Collection("todos"))

	tasks := []*pb.TodoObject{
		{
			Id:   "1",
			Task: "風呂掃除",
		},
		{
			Id:   "2",
			Task: "トイレ掃除",
		},
	}
	return &pb.TodoResponse{Todos: tasks}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
