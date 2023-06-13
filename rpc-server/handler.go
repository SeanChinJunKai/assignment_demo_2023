package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

var regex = "([A-Za-z]+):([A-Za-z]+)"

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	if err := checkSendRequest(req); err != nil {
		return nil, err
	}
	message := req.GetMessage()
	err := sendMessage(message.GetChat(), message.GetText(), message.GetSender())
	if err != nil {
		return nil, err
	}
	resp := rpc.NewSendResponse()
	resp.Code, resp.Msg = 0, "success"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	if err := checkPullRequest(req); err != nil {
		return nil, err
	}

	rpcMessages := make([]*rpc.Message, 0)
	limit := req.GetLimit()
	if limit == 0 {
		limit = 10
	}

	messages, err := pullMessage(req.GetChat(), limit+1, req.GetCursor(), req.GetReverse())
	if err != nil {
		return nil, err
	}
	var counter int64 = 0
	var hasMore = false
	var nextCursor int64 = 0
	for _, message := range messages {
		if counter+1 > int64(limit) {
			hasMore = true
			nextCursor = counter
			break
		}
		temp := &rpc.Message{
			Chat:     message.ChatId,
			Text:     message.Content,
			Sender:   message.Sender,
			SendTime: message.SendTime,
		}
		counter += 1
		rpcMessages = append(rpcMessages, temp)
	}
	resp := rpc.NewPullResponse()
	resp.Messages = rpcMessages
	resp.HasMore = &hasMore
	resp.NextCursor = &nextCursor
	resp.Code, resp.Msg = 0, "success"
	return resp, nil
}

func checkSendRequest(req *rpc.SendRequest) error {
	message := req.Message

	match, _ := regexp.MatchString(regex, message.Chat)
	if !match {
		err := fmt.Errorf("invalid Chat ID '%s', should be in the format of user1:user2", req.Message.GetChat())
		return err
	}

	if message.GetText() == "" {
		err := fmt.Errorf("message content cannot be empty")
		return err
	}

	if message.GetSender() == "" {
		err := fmt.Errorf("sender name cannot be empty")
		return err
	}

	req.Message.Chat = strings.ToLower(req.Message.Chat)
	req.Message.Sender = strings.ToLower(req.Message.Sender)

	participants := strings.Split(message.Chat, ":")

	if !contains(participants, message.GetSender()) {
		err := fmt.Errorf("sender is in the wrong chat room")
		return err
	}

	sort.Strings(participants)
	req.Message.Chat = strings.Join(participants, ":")

	return nil
}

func checkPullRequest(req *rpc.PullRequest) error {
	participants := strings.Split(req.GetChat(), ":")
	if len(participants) != 2 {
		err := fmt.Errorf("invalid Chat ID '%s', should be in the format of user1:user2", req.GetChat())
		return err
	}

	if req.Limit < 0 {
		err := fmt.Errorf("limit cannot be negative")
		return err
	}

	if req.Cursor < 0 {
		err := fmt.Errorf("cursor cannot be negative")
		return err
	}

	sort.Strings(participants)
	req.Chat = strings.Join(participants, ":")

	return nil
}

func contains(arr []string, element string) bool {
	for _, value := range arr {
		if value == element {
			return true
		}
	}
	return false
}
