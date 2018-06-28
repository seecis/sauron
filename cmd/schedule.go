// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

// scheduleCmd represents the schedule command
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Procrastinates jobs for you",
	Long:  `Schedule a task for a later execution`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		return
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}
