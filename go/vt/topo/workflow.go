package topo

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	workflowpb "github.com/youtube/vitess/go/vt/proto/workflow"
)

// This file provides the utility methods to save / retrieve workflows in the topology Backend.

const (
	workflowPath     = "/workflows/"
	workflowFilename = "Workflow"
)

func pathForWorkflow(uuid string) string {
	return workflowPath + uuid + "/" + workflowFilename
}

// WorkflowInfo is a meta struct that contains the version of a Workflow.
type WorkflowInfo struct {
	version Version
	*workflowpb.Workflow
}

// GetWorkflowNames returns the names of the existing
// workflows. They are sorted by uuid.
func (ts Server) GetWorkflowNames(ctx context.Context) ([]string, error) {
	entries, err := ts.ListDir(ctx, "global", workflowPath)
	switch err {
	case ErrNoNode:
		return nil, nil
	case nil:
		return entries, nil
	default:
		return nil, err
	}
}

// CreateWorkflow creates the given workflow, and returns the initial
// WorkflowInfo.
func (ts Server) CreateWorkflow(ctx context.Context, w *workflowpb.Workflow) (*WorkflowInfo, error) {
	// Pack the content.
	contents, err := proto.Marshal(w)
	if err != nil {
		return nil, err
	}

	// Save it.
	filePath := pathForWorkflow(w.Uuid)
	version, err := ts.Create(ctx, "global", filePath, contents)
	if err != nil {
		return nil, err
	}
	return &WorkflowInfo{
		version:  version,
		Workflow: w,
	}, nil
}

// GetWorkflow reads a workflow from the Backend.
func (ts Server) GetWorkflow(ctx context.Context, uuid string) (*WorkflowInfo, error) {
	// Read the file.
	filePath := pathForWorkflow(uuid)
	contents, version, err := ts.Get(ctx, "global", filePath)
	if err != nil {
		return nil, err
	}

	// Unpack the contents.
	w := &workflowpb.Workflow{}
	if err := proto.Unmarshal(contents, w); err != nil {
		return nil, err
	}

	return &WorkflowInfo{
		version:  version,
		Workflow: w,
	}, nil
}

// SaveWorkflow saves the WorkflowInfo object. If the version is not
// good any more, ErrBadVersion is returned.
func (ts Server) SaveWorkflow(ctx context.Context, wi *WorkflowInfo) error {
	// Pack the content.
	contents, err := proto.Marshal(wi.Workflow)
	if err != nil {
		return err
	}

	// Save it.
	filePath := pathForWorkflow(wi.Uuid)
	version, err := ts.Update(ctx, "global", filePath, contents, wi.version)
	if err != nil {
		return err
	}

	// Remember the new version.
	wi.version = version
	return nil
}

// DeleteWorkflow deletes the specified workflow.  After this, the
// WorkflowInfo object should not be used any more.
func (ts Server) DeleteWorkflow(ctx context.Context, wi *WorkflowInfo) error {
	filePath := pathForWorkflow(wi.Uuid)
	return ts.Delete(ctx, "global", filePath, wi.version)
}
