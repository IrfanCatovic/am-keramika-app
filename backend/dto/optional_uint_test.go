package dto

import (
	"encoding/json"
	"testing"
)

func TestOptionalUintUnmarshal(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		var req UpdateProductRequest
		err := json.Unmarshal([]byte(`{
			"name":"A",
			"categoryID":1,
			"unit":"kom",
			"salePrice":10,
			"stockQuantity":1
		}`), &req)
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if req.GroupID.Present {
			t.Fatalf("expected Present=false when groupID omitted")
		}
	})

	t.Run("explicit null", func(t *testing.T) {
		var req UpdateProductRequest
		err := json.Unmarshal([]byte(`{
			"name":"A",
			"categoryID":1,
			"groupID":null,
			"unit":"kom",
			"salePrice":10,
			"stockQuantity":1
		}`), &req)
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !req.GroupID.Present {
			t.Fatalf("expected Present=true for explicit null")
		}
		if req.GroupID.Value != nil {
			t.Fatalf("expected Value=nil for explicit null")
		}
	})

	t.Run("numeric value", func(t *testing.T) {
		var req UpdateProductRequest
		err := json.Unmarshal([]byte(`{
			"name":"A",
			"categoryID":1,
			"groupID":4,
			"unit":"kom",
			"salePrice":10,
			"stockQuantity":1
		}`), &req)
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !req.GroupID.Present || req.GroupID.Value == nil || *req.GroupID.Value != 4 {
			t.Fatalf("expected Present=true Value=4, got Present=%v Value=%v", req.GroupID.Present, req.GroupID.Value)
		}
	})
}
