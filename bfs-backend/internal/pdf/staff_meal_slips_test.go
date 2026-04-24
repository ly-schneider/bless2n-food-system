package pdf

import (
	"bytes"
	"testing"
)

func TestRenderStaffMealSlips_ProducesPDF(t *testing.T) {
	out, err := RenderStaffMealSlips(SlipInput{
		CampaignName: "Helferessen Samstag",
		QRPayload:    "CAMP:01937c7a-0000-7000-8000-000000000000",
		Products: []SlipProduct{
			{Name: "Pizza Margherita", Quantity: 1},
			{Name: "Getränk", Quantity: 1},
		},
		Count: 30,
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatalf("output missing PDF magic bytes; got first 8 bytes = %q", out[:8])
	}
	if len(out) < 2000 {
		t.Fatalf("PDF suspiciously small: %d bytes", len(out))
	}
}

func TestRenderStaffMealSlips_RejectsBadInput(t *testing.T) {
	if _, err := RenderStaffMealSlips(SlipInput{Count: 0, QRPayload: "x"}); err == nil {
		t.Fatal("expected error for count=0")
	}
	if _, err := RenderStaffMealSlips(SlipInput{Count: 1, QRPayload: ""}); err == nil {
		t.Fatal("expected error for empty payload")
	}
}
