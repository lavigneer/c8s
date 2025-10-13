package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

var pipelineConfigGVR = schema.GroupVersionResource{
	Group:    "c8s.io",
	Version:  "v1alpha1",
	Resource: "pipelineconfigs",
}

func getCommand(args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("resource type required (runs, configs)")
	}

	resourceType := fs.Arg(0)
	resourceName := ""
	if fs.NArg() > 1 {
		resourceName = fs.Arg(1)
	}

	switch resourceType {
	case "runs", "run", "pipelineruns", "pipelinerun":
		return getRuns(resourceName)
	case "configs", "config", "pipelineconfigs", "pipelineconfig":
		return getConfigs(resourceName)
	default:
		return fmt.Errorf("unknown resource type: %s. Available: runs, configs", resourceType)
	}
}

func getRuns(name string) error {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	ctx := context.Background()

	if name != "" {
		// Get specific run
		run, err := dynamicClient.Resource(pipelineRunGVR).Namespace(namespace).Get(
			ctx,
			name,
			metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to get PipelineRun: %w", err)
		}

		return printRunDetails(run)
	}

	// List all runs
	list, err := dynamicClient.Resource(pipelineRunGVR).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to list PipelineRuns: %w", err)
	}

	if len(list.Items) == 0 {
		fmt.Println("No PipelineRuns found")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tCONFIG\tCOMMIT\tBRANCH\tPHASE\tAGE")

	for _, item := range list.Items {
		spec, _, _ := unstructured.NestedMap(item.Object, "spec")
		status, _, _ := unstructured.NestedMap(item.Object, "status")

		configRef, _, _ := unstructured.NestedMap(spec, "pipelineConfigRef")
		configName, _, _ := unstructured.NestedString(configRef, "name")
		commit, _, _ := unstructured.NestedString(spec, "commit")
		branch, _, _ := unstructured.NestedString(spec, "branch")
		phase, _, _ := unstructured.NestedString(status, "phase")

		if phase == "" {
			phase = "Pending"
		}

		// Truncate commit to 7 chars
		if len(commit) > 7 {
			commit = commit[:7]
		}

		// Calculate age
		creationTimestamp := item.GetCreationTimestamp()
		age := time.Since(creationTimestamp.Time).Round(time.Second)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.GetName(),
			configName,
			commit,
			branch,
			phase,
			formatDuration(age),
		)
	}

	w.Flush()
	return nil
}

func getConfigs(name string) error {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	ctx := context.Background()

	if name != "" {
		// Get specific config
		config, err := dynamicClient.Resource(pipelineConfigGVR).Namespace(namespace).Get(
			ctx,
			name,
			metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to get PipelineConfig: %w", err)
		}

		return printConfigDetails(config)
	}

	// List all configs
	list, err := dynamicClient.Resource(pipelineConfigGVR).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to list PipelineConfigs: %w", err)
	}

	if len(list.Items) == 0 {
		fmt.Println("No PipelineConfigs found")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tREPOSITORY\tBRANCHES\tSTEPS\tAGE")

	for _, item := range list.Items {
		spec, _, _ := unstructured.NestedMap(item.Object, "spec")

		repository, _, _ := unstructured.NestedString(spec, "repository")
		branches, _, _ := unstructured.NestedStringSlice(spec, "branches")
		steps, _, _ := unstructured.NestedSlice(spec, "steps")

		branchesStr := strings.Join(branches, ",")
		if len(branchesStr) > 30 {
			branchesStr = branchesStr[:27] + "..."
		}

		creationTimestamp := item.GetCreationTimestamp()
		age := time.Since(creationTimestamp.Time).Round(time.Second)

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			item.GetName(),
			repository,
			branchesStr,
			len(steps),
			formatDuration(age),
		)
	}

	w.Flush()
	return nil
}

func printRunDetails(run *unstructured.Unstructured) error {
	spec, _, _ := unstructured.NestedMap(run.Object, "spec")
	status, _, _ := unstructured.NestedMap(run.Object, "status")

	configRef, _, _ := unstructured.NestedMap(spec, "pipelineConfigRef")
	configName, _, _ := unstructured.NestedString(configRef, "name")
	commit, _, _ := unstructured.NestedString(spec, "commit")
	branch, _, _ := unstructured.NestedString(spec, "branch")
	triggeredBy, _, _ := unstructured.NestedString(spec, "triggeredBy")
	triggeredAt, _, _ := unstructured.NestedString(spec, "triggeredAt")

	phase, _, _ := unstructured.NestedString(status, "phase")
	startTime, _, _ := unstructured.NestedString(status, "startTime")
	completionTime, _, _ := unstructured.NestedString(status, "completionTime")
	steps, _, _ := unstructured.NestedSlice(status, "steps")

	if phase == "" {
		phase = "Pending"
	}

	fmt.Printf("Name:            %s\n", run.GetName())
	fmt.Printf("Namespace:       %s\n", run.GetNamespace())
	fmt.Printf("Pipeline Config: %s\n", configName)
	fmt.Printf("Commit:          %s\n", commit)
	fmt.Printf("Branch:          %s\n", branch)
	fmt.Printf("Triggered By:    %s\n", triggeredBy)
	fmt.Printf("Triggered At:    %s\n", triggeredAt)
	fmt.Printf("Phase:           %s\n", phase)
	fmt.Printf("Start Time:      %s\n", startTime)
	fmt.Printf("Completion Time: %s\n", completionTime)

	if len(steps) > 0 {
		fmt.Println("\nSteps:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "  NAME\tPHASE\tJOB\tEXIT CODE\tDURATION")

		for _, stepObj := range steps {
			step := stepObj.(map[string]interface{})
			name, _ := step["name"].(string)
			stepPhase, _ := step["phase"].(string)
			jobName, _ := step["jobName"].(string)
			exitCode, _ := step["exitCode"].(float64)
			stepStartTime, _ := step["startTime"].(string)
			stepCompletionTime, _ := step["completionTime"].(string)

			var duration string
			if stepStartTime != "" && stepCompletionTime != "" {
				start, _ := time.Parse(time.RFC3339, stepStartTime)
				end, _ := time.Parse(time.RFC3339, stepCompletionTime)
				duration = formatDuration(end.Sub(start))
			}

			fmt.Fprintf(w, "  %s\t%s\t%s\t%.0f\t%s\n",
				name, stepPhase, jobName, exitCode, duration)
		}
		w.Flush()
	}

	return nil
}

func printConfigDetails(config *unstructured.Unstructured) error {
	spec, _, _ := unstructured.NestedMap(config.Object, "spec")

	repository, _, _ := unstructured.NestedString(spec, "repository")
	branches, _, _ := unstructured.NestedStringSlice(spec, "branches")
	timeout, _, _ := unstructured.NestedString(spec, "timeout")
	steps, _, _ := unstructured.NestedSlice(spec, "steps")

	fmt.Printf("Name:       %s\n", config.GetName())
	fmt.Printf("Namespace:  %s\n", config.GetNamespace())
	fmt.Printf("Repository: %s\n", repository)
	fmt.Printf("Branches:   %s\n", strings.Join(branches, ", "))
	fmt.Printf("Timeout:    %s\n", timeout)

	if len(steps) > 0 {
		fmt.Println("\nSteps:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "  NAME\tIMAGE\tCOMMANDS\tDEPENDS ON")

		for _, stepObj := range steps {
			step := stepObj.(map[string]interface{})
			name, _ := step["name"].(string)
			image, _ := step["image"].(string)
			commandsSlice, _ := step["commands"].([]interface{})
			dependsOnSlice, _ := step["dependsOn"].([]interface{})

			var commands []string
			for _, cmd := range commandsSlice {
				if cmdStr, ok := cmd.(string); ok {
					commands = append(commands, cmdStr)
				}
			}

			var dependsOn []string
			for _, dep := range dependsOnSlice {
				if depStr, ok := dep.(string); ok {
					dependsOn = append(dependsOn, depStr)
				}
			}

			commandsStr := strings.Join(commands, "; ")
			if len(commandsStr) > 40 {
				commandsStr = commandsStr[:37] + "..."
			}

			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
				name, image, commandsStr, strings.Join(dependsOn, ","))
		}
		w.Flush()
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
