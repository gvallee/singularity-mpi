// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package configparser

import (
	"fmt"
	"strings"

	"github.com/sylabs/singularity-mpi/internal/pkg/kv"
)

// Config represents the configuration of the tests to run
type Config struct {
	filename  string
	MPIImplem string            // Reference the MPI implementation, e.g., Open MPI, MPICH
	MpiMap    map[string]string // store the URL to download a specific version, the key being the version
}

func detectOpenMPIVersion(line string) string {
	if strings.Contains(line, "openmpi-") {
		if strings.Contains(line, ".tar") {
			endVersion := strings.Index(line, ".tar")
			startVersion := strings.Index(line, "openmpi-") + 8
			if startVersion != -1 && endVersion != -1 {
				return line[startVersion:endVersion]
			}
		}
	}

	// Could not detect anything
	return ""
}

func detectMPICHVersion(line string) string {
	if strings.Contains(line, "mpich-") {
		if strings.Contains(line, ".tar") {
			endVersion := strings.Index(line, ".tar")
			startVersion := strings.Index(line, "mpich-") + 6
			if startVersion != -1 && endVersion != -1 {
				return line[startVersion:endVersion]
			}
		}
	}

	// Could not detect anything
	return ""
}

func detectIntelMPIVersion(line string) string {
	if strings.Contains(line, "l_mpi_") {
		if strings.Contains(line, ".tar") {
			endVersion := strings.Index(line, ".tar")
			startVersion := strings.Index(line, "l_mpi_")
			if startVersion != -1 && endVersion != -1 {
				return line[startVersion:endVersion]
			}
		}
	}

	// Could not detect anything
	return ""
}

func detectMpiImplem(line string) (string, string) {
	// The line that is passed in has a format similar to: https://download.open-mpi.org/release/open-mpi/v3.0/openmpi-3.0.4.tar.bz2
	ompiVer := detectOpenMPIVersion(line)
	if ompiVer != "" {
		return "openmpi", ompiVer
	}

	mpichVer := detectMPICHVersion(line)
	if mpichVer != "" {
		return "mpich", mpichVer
	}

	intelVer := detectIntelMPIVersion(line)
	if intelVer != "" {
		return "intel", intelVer
	}

	return "", ""
}

// Parse go through the configuration file to load the associated configuration
func Parse(file string) (*Config, error) {
	/*
		// Open the config file
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("cannot open file: %s", err)
		}
		defer f.Close()
	*/

	// Allocate a new config data structure
	config := new(Config)
	config.filename = file

	config.MpiMap = make(map[string]string)

	kvs, err := kv.LoadKeyValueConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %s", file, err)
	}

	for _, kv := range kvs {
		// If we did not detect the MPI implementation yet, we try to detect it
		implem, version := detectMpiImplem(kv.Value)
		if implem == "" || version == "" {
			return nil, fmt.Errorf("cannot detect the MPI implementation from %s", kv.Value)
		}

		// If we did not detect the MPI implementation yet, we save it
		if config.MPIImplem == "" {
			config.MPIImplem = implem
		} else if config.MPIImplem != implem {
			return nil, fmt.Errorf("Detected two implementations of MPI (%s and %s)", config.MPIImplem, implem)
		}

		config.MpiMap[version] = kv.Value
	}

	return config, nil
}
