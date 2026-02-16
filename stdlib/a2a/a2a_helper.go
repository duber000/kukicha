package a2a

import (
	"context"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/a2a"
)

// Send executes a blocking request and returns the task result.
func Send(req Request) (Task, error) {
	return sendRequest(req.agent, req.text, req.contextID)
}

// Stream executes a streaming request with callbacks.
func Stream(req Request) (Task, error) {
	return streamRequest(req.agent, req.text, req.contextID, req.onText, req.onStatus)
}

// Ask is a one-shot convenience: send text and get the reply text back.
func Ask(agent Agent, text string) (string, error) {
	task, err := sendRequest(agent, text, "")
	if err != nil {
		return "", err
	}
	return task.Text, nil
}

// GetTask queries a task by ID from the agent.
func GetTask(agent Agent, taskID string) (Task, error) {
	t, err := agent.Client.GetTask(context.Background(), &a2a.TaskQueryParams{
		ID: a2a.TaskID(taskID),
	})
	if err != nil {
		return Task{}, fmt.Errorf("a2a get task: %w", err)
	}
	return taskFromA2A(t), nil
}

// Cancel cancels a task by ID.
func Cancel(agent Agent, taskID string) (Task, error) {
	t, err := agent.Client.CancelTask(context.Background(), &a2a.TaskIDParams{
		ID: a2a.TaskID(taskID),
	})
	if err != nil {
		return Task{}, fmt.Errorf("a2a cancel: %w", err)
	}
	return taskFromA2A(t), nil
}

// sendRequest sends a blocking message to an agent and returns a simplified Task.
func sendRequest(agent Agent, text string, contextID string) (Task, error) {
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.TextPart{Text: text})
	if contextID != "" {
		msg.ContextID = contextID
	}
	resp, err := agent.Client.SendMessage(context.Background(), &a2a.MessageSendParams{
		Message: msg,
	})
	if err != nil {
		return Task{}, fmt.Errorf("a2a send: %w", err)
	}
	return resultToTask(resp), nil
}

// streamRequest sends a streaming message to an agent, dispatching to callbacks.
func streamRequest(agent Agent, text string, contextID string, onText TextHandler, onStatus StatusHandler) (Task, error) {
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.TextPart{Text: text})
	if contextID != "" {
		msg.ContextID = contextID
	}

	var result Task
	for event, err := range agent.Client.SendStreamingMessage(context.Background(), &a2a.MessageSendParams{
		Message: msg,
	}) {
		if err != nil {
			return result, fmt.Errorf("a2a stream: %w", err)
		}
		switch e := event.(type) {
		case *a2a.TaskStatusUpdateEvent:
			if onStatus != nil {
				statusMsg := ""
				if e.Status.Message != nil {
					statusMsg = extractPartsText(e.Status.Message.Parts)
				}
				onStatus(StatusUpdate{
					TaskID:  string(e.TaskID),
					State:   string(e.Status.State),
					Message: statusMsg,
					Final:   e.Final,
				})
			}
			result.ID = string(e.TaskID)
			result.ContextID = e.ContextID
			result.State = string(e.Status.State)
		case *a2a.TaskArtifactUpdateEvent:
			if e.Artifact != nil {
				artText := extractPartsText(e.Artifact.Parts)
				if onText != nil && artText != "" {
					onText(artText)
				}
				result.Artifacts = append(result.Artifacts, Artifact{
					Name: e.Artifact.Name,
					Text: artText,
				})
			}
			result.ID = string(e.TaskID)
			result.ContextID = e.ContextID
		case *a2a.Task:
			result = taskFromA2A(e)
		case *a2a.Message:
			msgText := extractPartsText(e.Parts)
			if onText != nil && msgText != "" {
				onText(msgText)
			}
			result.Text = msgText
		}
	}
	return result, nil
}

// resultToTask converts a SendMessageResult (union of *Task or *Message) to our simplified Task.
func resultToTask(result a2a.SendMessageResult) Task {
	switch r := result.(type) {
	case *a2a.Task:
		return taskFromA2A(r)
	case *a2a.Message:
		return Task{
			ID:        string(r.TaskID),
			ContextID: r.ContextID,
			Text:      extractPartsText(r.Parts),
		}
	default:
		return Task{}
	}
}

// taskFromA2A converts an a2a.Task to our simplified Task type.
func taskFromA2A(t *a2a.Task) Task {
	result := Task{
		ID:        string(t.ID),
		ContextID: t.ContextID,
		State:     string(t.Status.State),
	}

	for _, art := range t.Artifacts {
		artText := extractPartsText(art.Parts)
		result.Artifacts = append(result.Artifacts, Artifact{
			Name: art.Name,
			Text: artText,
		})
	}

	var texts []string
	if t.Status.Message != nil {
		if msg := extractPartsText(t.Status.Message.Parts); msg != "" {
			texts = append(texts, msg)
		}
	}
	for _, art := range result.Artifacts {
		if art.Text != "" {
			texts = append(texts, art.Text)
		}
	}
	result.Text = strings.Join(texts, "\n")

	return result
}

// extractPartsText concatenates all TextPart content from a list of parts.
func extractPartsText(parts []a2a.Part) string {
	var texts []string
	for _, p := range parts {
		if tp, ok := p.(a2a.TextPart); ok {
			texts = append(texts, tp.Text)
		}
	}
	return strings.Join(texts, "")
}
