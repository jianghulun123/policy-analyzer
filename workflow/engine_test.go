package workflow

import "testing"

func TestGetResultRunningIncludesStepNameShape(t *testing.T) {
	engine := &Engine{
		workflows: map[string]*WorkflowDefinition{
			"wf-1": {
				ID: "wf-1",
				Steps: []StepDef{
					{ID: "step-a", Name: "步骤A"},
				},
			},
		},
		executions: map[string]*ExecutionSession{
			"exec-1": {
				ID:         "exec-1",
				WorkflowID: "wf-1",
				Status:     "running",
				Context: &ExecutionContext{
					StepResults: map[string]SkillOutput{
						"step-a": {
							Data: map[string]interface{}{"report": "ok"},
						},
					},
				},
			},
		},
	}

	result, err := engine.GetResult("exec-1")
	if err != nil {
		t.Fatalf("GetResult failed: %v", err)
	}

	step, ok := result.StepResults["step-a"]
	if !ok {
		t.Fatal("expected running step result")
	}
	if step.Name != "步骤A" {
		t.Fatalf("expected step name 步骤A, got %q", step.Name)
	}
	if step.Status != "completed" {
		t.Fatalf("expected completed status, got %q", step.Status)
	}
	if step.Error != "" {
		t.Fatalf("expected empty error, got %q", step.Error)
	}
	if step.Duration != 0 {
		t.Fatalf("expected zero duration, got %d", step.Duration)
	}
}
