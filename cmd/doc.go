/*
Copyright © 2019 Ilya V. Logounov <ilya@logounov.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"github.com/Logunov/heydevops/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docCmd represents the doc command
var (
	docCmd = &cobra.Command{
		Use:    "doc <man|md>",
		Short:  "",
		Long:   ``,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := doc.GenMarkdownTree(rootCmd, "./")
			helpers.CheckPanic(err)
		},
	}
)

func init() {
	rootCmd.AddCommand(docCmd)
}
