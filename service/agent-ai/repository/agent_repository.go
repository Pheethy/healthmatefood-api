package repository

import (
	"bytes"
	"context"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/models"
	"healthmatefood-api/service/agent-ai"
	"html/template"
	"log"

	"github.com/tmc/langchaingo/llms"
)

type agentAIRepository struct {
	cfg             config.IAgentConfig
	digitalOceanLLM *digitalOceanLLM
}

func NewAgentAIRepository(cfg config.IAgentConfig) agent.IAgentAIRepository {
	return &agentAIRepository{
		cfg:             cfg,
		digitalOceanLLM: NewDigitalOceanLLM(cfg.AgentEndpoint(), cfg.AgentAccessKey()),
	}
}

func (r *agentAIRepository) GenerateMealsPlan(ctx context.Context, user *models.User) (string, error) {
	path := "templates/user_info.txt"
	tmpl := template.Must(template.ParseFiles(path))
	var prompt bytes.Buffer
	if err := tmpl.Execute(&prompt, user.UserInfo); err != nil {
		return "", err
	}

	log.Println("prompt", prompt.String())
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "คุณเป็นผู้เชี่ยวชาญที่ให้คำตอบเป็นภาษาไทยเท่านั้น"}},
		},
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: `
		    Variable Definitions
		    [age]: The individual’s age in years.
		    [gender]: The individual’s gender (e.g., male, female, or non-binary).
		    [medical_condition]: A specific health condition that impacts dietary choices (e.g., diabetes, hypertension, renal impairment).
		    [calories]: A specific calorie limit (e.g., daily total, per meal).
		    [activity_level]: Physical activity intensity, such as sedentary, moderate, or active.
		    [activity_level_active]: A lifestyle involving frequent physical activity.
		    [food_or_ingredients]: The name of a dish, food item, or a list of ingredients.
		    [food_log]: A record of meals consumed.
		    [food_image]: An uploaded image of a meal or pantry contents.
		    [cuisine_type]: The style or origin of cuisine (e.g., Italian, Thai).
		    [carbohydrate_limit]: The maximum allowable carbohydrate intake in grams.
		    [calories_burned]: Calories burned during physical activity.
		    [nutritional_focus]: A dietary priority such as high-protein, low-fat, or fiber-rich.
		    [weight_goal]: Desired weight change in kilograms.
		    [timeframe]: The time period to achieve the weight goal (e.g., in weeks).
		    [food_preferences_like]: Foods that the individual enjoys.
		    [food_preferences_dislike]: Foods that the individual dislikes.

		    General Food and Recipe Recommendations
		    "Recommend meal plans for a [age]-year-old [gender] with [medical_condition] requiring a daily intake of [calories] calories."
		    "Provide a list of recipes under [calories] calories, suitable for someone with [medical_condition]."
		    "What are some high-protein, low-sodium dinner options for a [age]-year-old [gender] with [medical_condition]?"

		    Nutritional Analysis and Recommendations
		    "Calculate the daily energy requirement for a [age]-year-old [gender] with a weight of [weight] kg, height of [height] cm, and [activity_level], considering their [medical_condition]."
		    "Analyze the nutritional content of this meal: [food_or_ingredients], and suggest healthier substitutions."
		    "How many calories are in a serving of [food_or_ingredients]?"
		    "Provide the macronutrient breakdown (carbs, protein, fat) for [food_or_ingredients]."

		    Interactive Food Tracking
		    "Log this meal: [food_or_ingredients], and calculate its calorie and nutritional value."
		    "Based on today’s meals: [food_log], suggest a balanced dinner to meet my remaining nutritional goals."
		    "Analyze this image of my meal: [food_image]. Estimate its calories and macronutrients."`}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt.String()}},
		},
	}

	resp, err := r.digitalOceanLLM.GenerateContent(ctx, messages)
	if err != nil {
		log.Println("err", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return resp.Choices[0].Content, nil
}

func (r *agentAIRepository) ConversationWithChat(ctx context.Context, prompt string) (string, error) {
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "คุณเป็นผู้เชี่ยวชาญที่ให้คำตอบเป็นภาษาไทยเท่านั้น"}},
		},
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "ตอบเป็นแต่ละวันทำอะไร เวลาไหน ด้วยนะ"}},
		},
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "ตอบเป็น json format เท่านั้น"}},
		},
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: `
                JSON FORMAT: 
                {
                    "plan": {
                        "desination": "ระยอง",
                        "start_date": "2023-01-01",
                        "end_date": "2023-01-02",
                        "total_day": 2,
                        "itinerary": [
                            {
                                "day": 1,
                                "activity": ["เช้า: ดำน้ำ", "เที่ยง: กินอาหารทะเล", "เย็น: ไปดู pingpong show"]
                            },
                            {
                                "day": 2,
                                "activity": ["เช้า: เที่ยวเกาะล้าน", "เที่ยง: กินอาหารบนเกาะ", "เย็น: รับประทานอาหารเย็นและชมการแสดง"]
                            }
                        ]
                    }
                }
            `}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
	}

	response, err := r.digitalOceanLLM.GenerateContent(ctx, messages)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return response.Choices[0].Content, nil
}
