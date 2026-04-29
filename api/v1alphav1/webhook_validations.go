// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateSpecOvercommit(class OvercommitClass) error {
	if class.Spec.CpuOvercommit <= 0 || class.Spec.CpuOvercommit > 1 {
		return errors.New("Error: cpuOvercommit must be greater than 0 and equal or lower than 1, failed creating " + class.Name + " class ")
	}
	if class.Spec.MemoryOvercommit <= 0 || class.Spec.MemoryOvercommit > 1 {
		return errors.New("Error: memoryOvercommit must be greater than 0 and equal or lower than 1, failed creating " + class.Name + " class ")
	}
	return nil
}

func checkDecimals(class OvercommitClass) error {
	cpu := class.Spec.CpuOvercommit
	memory := class.Spec.MemoryOvercommit
	const precision = 10000 // 10^4
	roundedCpu := math.Round(cpu*precision) / precision

	if math.Abs(cpu-roundedCpu) > 1e-9 {
		return errors.New("the CPU value must have 4 decimals max")
	}

	roundedMemory := math.Round(memory*precision) / precision
	if math.Abs(memory-roundedMemory) > 1e-9 {
		return errors.New("the memory value must have 4 decimals max")
	}
	return nil
}

func isClassDefault(class OvercommitClass, client client.Client) error {
	// Create a context for the client
	ctx := context.TODO()

	// List all OvercommitClasses
	var overcommitClassList OvercommitClassList
	err := client.List(ctx, &overcommitClassList)
	if err != nil {
		return fmt.Errorf("error listing OvercommitClasses: %w", err)
	}

	var existsDefault = false
	if class.Spec.IsDefault {
		for _, item := range overcommitClassList.Items {
			if item.Spec.IsDefault {
				existsDefault = true
			}
		}
	}

	if existsDefault {
		return fmt.Errorf("error: only one OvercommitClass can be default, failed creating %s class", class.Name)
	}

	return nil
}

func checkIsRegexValid(regex string) error {
	// Limit regex length to prevent ReDoS (catastrophic backtracking)
	const maxRegexLen = 512
	if len(regex) > maxRegexLen {
		return fmt.Errorf("regex is too long (%d chars), maximum allowed is %d", len(regex), maxRegexLen)
	}
	_, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("invalid regex for excludedNamespaces: %w", err)
	}
	return nil
}
