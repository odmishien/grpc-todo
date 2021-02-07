package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"cloud.google.com/go/firestore"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/odmishien/grpctodo/config"
	"github.com/odmishien/grpctodo/interceptor"
	pb "github.com/odmishien/grpctodo/todo"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port      = ":50051"
	projectid = "grpctodo"
)

type server struct {
	pb.UnimplementedTodoServiceServer
}

func (s *server) AddTodo(ctx context.Context, in *pb.AddTodoParams) (*pb.TodoObject, error) {
	uid := getUserID(ctx)
	if uid == "" {
		return nil, errors.New("missing userid")
	}
	task := in.GetTask()
	fmt.Printf("task: %s\n", task)
	if task == "" {
		return nil, errors.New("invalid argument")
	}
	fs, err := firestore.NewClient(ctx, projectid)
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("failed to init firestore client")
	}
	defer fs.Close()
	t, _, err := fs.Collection("todos").Add(ctx, map[string]interface{}{
		"task":   task,
		"userId": uid,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Printf("%#v\n", t)
	return &pb.TodoObject{Id: t.ID, Task: task, UserId: uid}, nil
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
	uid := getUserID(ctx)
	if uid == "" {
		return nil, errors.New("missing userid")
	}
	fs, err := firestore.NewClient(ctx, projectid)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	tasks := make([]*pb.TodoObject, 0)
	iter := fs.Collection("todos").Where("userId", "==", uid).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data := doc.Data()
		tasks = append(tasks, &pb.TodoObject{
			Id:     doc.Ref.ID,
			Task:   data["task"].(string),
			UserId: data["userId"].(string),
		})
	}

	return &pb.TodoResponse{Todos: tasks}, nil
}

func getUserID(ctx context.Context) string {
	userID, ok := ctx.Value(config.UserKey).(string)
	if !ok {
		panic("userID missing")
	}
	return userID
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				interceptor.NewAuthInterceptor(),
			),
		),
	}
	s := grpc.NewServer(opts...)
	pb.RegisterTodoServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
