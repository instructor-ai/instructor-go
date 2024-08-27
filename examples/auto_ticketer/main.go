/*
 *  Original example in Python: https://github.com/jxnl/instructor/blob/11125a7c831a26e2a4deaef4129f2b4845a7e079/examples/auto-ticketer/run.py
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type PriorityEnum string

const (
	High   PriorityEnum = "High"
	Medium PriorityEnum = "Medium"
	Low    PriorityEnum = "Low"
)

type Subtask struct {
	ID   int    `json:"id"      jsonschema:"title=unique identifier for the subtask,description=Unique identifier for the subtask"`
	Name string `json:"name"    jsonschema:"title=name of the subtask,description=Informative title of the subtask"`
}

type Ticket struct {
	ID           int          `json:"id"                        jsonschema:"title=unique identifier for the ticket,description=Unique identifier for the ticket"`
	Name         string       `json:"name"                      jsonschema:"title=name of the task,description=Title of the task"`
	Description  string       `json:"description"               jsonschema:"title=description of the task,description=Detailed description of the task"`
	Priority     PriorityEnum `json:"priority"                  jsonschema:"title=priority level,description=Priority level"`
	Assignees    []string     `json:"assignees"                 jsonschema:"title=list of users assigned to the task,description=List of users assigned to the task"`
	Subtasks     []Subtask    `json:"subtasks"        jsonschema:"title=list of subtasks associated with the main task,description=List of subtasks associated with the main task"`
	Dependencies []int        `json:"dependencies"    jsonschema:"title=list of ticket IDs that this ticket depends on,description=List of ticket IDs that this ticket depends on"`
}

type ActionItems struct {
	Tickets []Ticket `json:"tickets"`
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSONStrict),
		instructor.WithMaxRetries(0),
	)

	transcript := `
Alice: Hey team, we have several critical tasks we need to tackle for the upcoming release. First, we need to work on improving the authentication system. It's a top priority.

Bob: Got it, Alice. I can take the lead on the authentication improvements. Are there any specific areas you want me to focus on?

Alice: Good question, Bob. We need both a front-end revamp and back-end optimization. So basically, two sub-tasks.

Carol: I can help with the front-end part of the authentication system.

Bob: Great, Carol. I'll handle the back-end optimization then.

Alice: Perfect. Now, after the authentication system is improved, we have to integrate it with our new billing system. That's a medium priority task.

Carol: Is the new billing system already in place?

Alice: No, it's actually another task. So it's a dependency for the integration task. Bob, can you also handle the billing system?

Bob: Sure, but I'll need to complete the back-end optimization of the authentication system first, so it's dependent on that.

Alice: Understood. Lastly, we also need to update our user documentation to reflect all these changes. It's a low-priority task but still important.

Carol: I can take that on once the front-end changes for the authentication system are done. So, it would be dependent on that.

Alice: Sounds like a plan. Let's get these tasks modeled out and get started.
`

	var actionItems ActionItems
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini20240718,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "The following is a transcript of a meeting between a manager and their team. The manager is assigning tasks to their team members and creating action items for them to complete.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Create the action items for the following transcript: %s", transcript),
				},
			},
		},
		&actionItems,
	)
	_ = resp // sends back original response so no information loss from original API
	if err != nil {
		log.Fatal(err)
	}

	prettyJSON, err := json.MarshalIndent(actionItems, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(prettyJSON))
}
