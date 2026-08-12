package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/pathguard"
)

func Sender(root string, arm model.Arm, fixture model.Fixture) (string, error) {
	common, err := read(root, "prompts/common/sender.md")
	if err != nil {
		return "", err
	}
	carrier, err := read(root, arm.CarrierPrompt)
	if err != nil {
		return "", err
	}
	policy, err := policyText(root, arm)
	if err != nil {
		return "", err
	}
	public, err := marshalPublic(fixture.SenderPublic())
	if err != nil {
		return "", err
	}
	return join(common, carrier, policy, "# Sender task data\n\n<sender_task_json>\n"+public+"\n</sender_task_json>"), nil
}

func Receiver(root string, arm model.Arm, fixture model.Fixture, handoff string) (string, error) {
	common, err := read(root, "prompts/common/receiver.md")
	if err != nil {
		return "", err
	}
	carrier, err := read(root, arm.CarrierPrompt)
	if err != nil {
		return "", err
	}
	policy, err := policyText(root, arm)
	if err != nil {
		return "", err
	}
	public, err := marshalPublic(fixture.ReceiverPublic())
	if err != nil {
		return "", err
	}
	return join(
		common,
		"# Handoff interpretation\n\nThe sender used the following carrier rules. Interpret the handoff accordingly.\n\n"+carrier,
		policy,
		"# Receiver task shell\n\nSource contents are intentionally absent. Use the sender handoff for derived facts and the task shell only for opaque ID allowlists and immutable authority.\n\n<receiver_task_json>\n"+public+"\n</receiver_task_json>",
		"# Exact sender handoff\n\n<sender_handoff>\n"+handoff+"\n</sender_handoff>",
	), nil
}

func marshalPublic(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	text := string(data)
	for _, forbidden := range []string{`"gold"`, `"class"`, `"title"`} {
		if strings.Contains(text, forbidden) {
			return "", fmt.Errorf("public fixture leaked hidden field %s", forbidden)
		}
	}
	return text, nil
}

func policyText(root string, arm model.Arm) (string, error) {
	if arm.PolicyPrompt == nil {
		return "", nil
	}
	return read(root, *arm.PolicyPrompt)
}

func read(root, rel string) (string, error) {
	path, err := pathguard.ResolveRegular(root, rel)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func join(parts ...string) string {
	var nonempty []string
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			nonempty = append(nonempty, strings.TrimSpace(part))
		}
	}
	return strings.Join(nonempty, "\n\n") + "\n"
}
