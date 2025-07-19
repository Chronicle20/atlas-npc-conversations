package conversation

import (
	"atlas-npc-conversations/saga"
	"context"
	"testing"

	"github.com/Chronicle20/atlas-constants/field"
	"github.com/Chronicle20/atlas-constants/world"
	"github.com/Chronicle20/atlas-tenant"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestExecutor creates an OperationExecutorImpl for testing
func createTestExecutor() *OperationExecutorImpl {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	
	// Create a test context with tenant
	ctx := context.Background()
	t := tenant.Model{} // Using zero-value tenant for testing
	
	return &OperationExecutorImpl{
		l:     logger,
		ctx:   ctx,
		t:     t,
		sagaP: nil, // We'll mock this if needed
	}
}

// createTestField creates a test field for operations
func createTestFieldForExecutor() field.Model {
	return field.NewBuilder(world.Id(1), 1, 100000).Build()
}

// TestOperationExecutor_IncreaseBuddyCapacity tests the buddy capacity operation
func TestOperationExecutor_IncreaseBuddyCapacity(t *testing.T) {
	tests := []struct {
		name            string
		params          map[string]string
		expectError     bool
		expectedError   string
		expectedAmount  int
	}{
		{
			name: "Valid amount parameter",
			params: map[string]string{
				"amount": "5",
			},
			expectError:    false,
			expectedAmount: 5,
		},
		{
			name: "Valid large amount",
			params: map[string]string{
				"amount": "50",
			},
			expectError:    false,
			expectedAmount: 50,
		},
		{
			name: "Valid maximum amount",
			params: map[string]string{
				"amount": "255",
			},
			expectError:    false,
			expectedAmount: 255,
		},
		{
			name: "Valid minimum amount",
			params: map[string]string{
				"amount": "1",
			},
			expectError:    false,
			expectedAmount: 1,
		},
		{
			name:          "Missing amount parameter",
			params:        map[string]string{},
			expectError:   true,
			expectedError: "missing amount parameter",
		},
		{
			name: "Invalid amount - zero",
			params: map[string]string{
				"amount": "0",
			},
			expectError:   true,
			expectedError: "amount [0] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name: "Invalid amount - negative",
			params: map[string]string{
				"amount": "-5",
			},
			expectError:   true,
			expectedError: "amount [-5] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name: "Invalid amount - too large",
			params: map[string]string{
				"amount": "256",
			},
			expectError:   true,
			expectedError: "amount [256] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name: "Invalid amount - not a number",
			params: map[string]string{
				"amount": "not_a_number",
			},
			expectError:   true,
			expectedError: "value [not_a_number] for parameter [amount] is not a valid integer",
		},
		{
			name: "Invalid amount - empty string",
			params: map[string]string{
				"amount": "",
			},
			expectError:   true,
			expectedError: "value [] for parameter [amount] is not a valid integer",
		},
		{
			name: "Invalid amount - decimal",
			params: map[string]string{
				"amount": "5.5",
			},
			expectError:   true,
			expectedError: "value [5.5] for parameter [amount] is not a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := createTestExecutor()
			testField := createTestFieldForExecutor()
			characterId := uint32(12345)

			// Create operation with test parameters
			operation := OperationModel{
				operationType: "increase_buddy_capacity",
				params:        tt.params,
			}

			// Call createStepForOperation to test parameter extraction and validation
			stepId, status, action, payload, err := executor.createStepForOperation(testField, characterId, operation)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, stepId)
				assert.Empty(t, status)
				assert.Empty(t, action)
				assert.Nil(t, payload)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "increase_buddy_capacity-12345", stepId)
				assert.Equal(t, saga.Pending, status)
				assert.Equal(t, saga.IncreaseBuddyCapacity, action)
				
				// Verify payload structure
				require.IsType(t, saga.IncreaseBuddyCapacityPayload{}, payload)
				buddyPayload := payload.(saga.IncreaseBuddyCapacityPayload)
				assert.Equal(t, characterId, buddyPayload.CharacterId)
				assert.Equal(t, testField.WorldId(), world.Id(buddyPayload.WorldId))
				assert.Equal(t, byte(tt.expectedAmount), buddyPayload.NewCapacity)
			}
		})
	}
}

// TestOperationExecutor_IncreaseBuddyCapacity_WithContextReferences tests context value resolution
func TestOperationExecutor_IncreaseBuddyCapacity_WithContextReferences(t *testing.T) {
	executor := createTestExecutor()
	testField := createTestFieldForExecutor()
	characterId := uint32(12345)

	// Create conversation context with test values
	ctx := ConversationContext{
		field:        testField,
		characterId:  characterId,
		npcId:        9001,
		currentState: "test_state",
		conversation: Model{}, // Empty for this test
		context: map[string]string{
			"buddyIncrease": "10",
			"maxIncrease":   "25",
		},
	}

	// Store context in registry
	GetRegistry().SetContext(executor.t, characterId, ctx)

	tests := []struct {
		name            string
		amountParam     string
		expectedAmount  int
		expectError     bool
		expectedError   string
	}{
		{
			name:           "Valid context reference",
			amountParam:    "context.buddyIncrease",
			expectedAmount: 10,
			expectError:    false,
		},
		{
			name:           "Valid context reference - different value",
			amountParam:    "context.maxIncrease",
			expectedAmount: 25,
			expectError:    false,
		},
		{
			name:          "Invalid context reference - missing key",
			amountParam:   "context.nonExistentKey",
			expectError:   true,
			expectedError: "context key [nonExistentKey] not found",
		},
		{
			name:          "Invalid context reference - malformed",
			amountParam:   "context.",
			expectError:   true,
			expectedError: "context key [] not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create operation with context reference
			operation := OperationModel{
				operationType: "increase_buddy_capacity",
				params: map[string]string{
					"amount": tt.amountParam,
				},
			}

			// Call createStepForOperation
			stepId, status, action, payload, err := executor.createStepForOperation(testField, characterId, operation)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "increase_buddy_capacity-12345", stepId)
				assert.Equal(t, saga.Pending, status)
				assert.Equal(t, saga.IncreaseBuddyCapacity, action)
				
				// Verify payload
				require.IsType(t, saga.IncreaseBuddyCapacityPayload{}, payload)
				buddyPayload := payload.(saga.IncreaseBuddyCapacityPayload)
				assert.Equal(t, characterId, buddyPayload.CharacterId)
				assert.Equal(t, testField.WorldId(), world.Id(buddyPayload.WorldId))
				assert.Equal(t, byte(tt.expectedAmount), buddyPayload.NewCapacity)
			}
		})
	}

	// Clean up
	GetRegistry().ClearContext(executor.t, characterId)
}

// TestOperationExecutor_IncreaseBuddyCapacity_EdgeCases tests boundary conditions
func TestOperationExecutor_IncreaseBuddyCapacity_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		amountStr     string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Boundary value - exactly 1",
			amountStr:   "1",
			expectError: false,
		},
		{
			name:        "Boundary value - exactly 255",
			amountStr:   "255",
			expectError: false,
		},
		{
			name:          "Just below minimum",
			amountStr:     "0",
			expectError:   true,
			expectedError: "amount [0] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name:          "Just above maximum",
			amountStr:     "256",
			expectError:   true,
			expectedError: "amount [256] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name:          "Large negative number",
			amountStr:     "-1000",
			expectError:   true,
			expectedError: "amount [-1000] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name:          "Very large positive number",
			amountStr:     "9999",
			expectError:   true,
			expectedError: "amount [9999] for increase_buddy_capacity operation must be between 1 and 255",
		},
		{
			name:          "Leading zeros - valid",
			amountStr:     "005",
			expectError:   false, // Should parse as 5
		},
		{
			name:          "Whitespace in number",
			amountStr:     " 10 ",
			expectError:   true,
			expectedError: "value [ 10 ] for parameter [amount] is not a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := createTestExecutor()
			testField := createTestFieldForExecutor()
			characterId := uint32(12345)

			operation := OperationModel{
				operationType: "increase_buddy_capacity",
				params: map[string]string{
					"amount": tt.amountStr,
				},
			}

			_, _, _, _, err := executor.createStepForOperation(testField, characterId, operation)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}