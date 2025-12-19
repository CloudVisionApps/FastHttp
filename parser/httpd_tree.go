package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseFileTree parses an Apache config file into a tree structure
func (p *ApacheHttpdParser) parseFileTree(filePath string) (*ConfigNode, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}
	originalBaseDir := p.baseDir
	p.baseDir = filepath.Dir(absPath)
	defer func() { p.baseDir = originalBaseDir }()

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	root := NewConfigNode("root", "", nil)
	currentNode := root
	stack := []*ConfigNode{root} // Stack to track nested blocks

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		originalLine := strings.TrimSpace(scanner.Text())
		line := originalLine

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Parse directive
		directive, args := p.parseDirective(line)
		if directive == "" {
			continue
		}

		// Handle closing tags
		if strings.HasPrefix(originalLine, "</") {
			// Pop from stack (go back to parent)
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
				currentNode = stack[len(stack)-1]
			}
			continue
		}

		// Handle IfModule blocks (skip processing their contents)
		if directive == "<IfModule" || strings.HasPrefix(directive, "<If") {
			// Create IfModule node but skip processing its contents
			ifModuleNode := NewConfigNode("IfModule", directive, args)
			currentNode.AddChild(ifModuleNode)
			stack = append(stack, ifModuleNode)
			currentNode = ifModuleNode
			continue
		}
		
		// Check if we're closing an IfModule
		if directive == "IfModule" && strings.HasPrefix(originalLine, "</") {
			// Pop from stack
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
				currentNode = stack[len(stack)-1]
			}
			continue
		}

		// Handle block directives (opening tags)
		if strings.HasPrefix(originalLine, "<") {
			blockType := directive
			
			// Create new block node
			blockNode := NewConfigNode(blockType, directive, args)
			currentNode.AddChild(blockNode)
			
			// Push to stack
			stack = append(stack, blockNode)
			currentNode = blockNode
			continue
		}

		// Handle simple directives
		// Add directive to current node (even if inside IfModule - we'll extract them later)
		currentNode.AddDirective(directive, args)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return root, nil
}
