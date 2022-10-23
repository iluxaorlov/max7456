package main

import (
	"errors"
	"github.com/iluxaorlov/max7456/internal/converter/mcm"
	"github.com/spf13/cobra"
)

func main() {
	command := &cobra.Command{
		Use:   "max7456",
		Short: "Allows you to convert *.mcm font file into an image collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path, err := cmd.Flags().GetString("decode"); err != nil {
				return err
			} else if path != "" {
				return mcm.NewConverter().Decode(path)
			}

			if name, err := cmd.Flags().GetString("encode"); err != nil {
				return err
			} else if name != "" {
				return mcm.NewConverter().Encode(name)
			}

			return errors.New("required at least one flag")
		},
	}

	command.Flags().StringP("decode", "d", "", "path to *.mcm file")
	command.Flags().StringP("encode", "e", "", "path to directory with images")

	command.Execute()
}
