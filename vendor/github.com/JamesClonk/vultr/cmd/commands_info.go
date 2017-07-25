package cmd

import (
	"log"

	"github.com/jawher/mow.cli"
)

func accountInfo(cmd *cli.Cmd) {
	cmd.Action = func() {
		info, err := GetClient().GetAccountInfo()
		if err != nil {
			log.Fatal(err)
		}

		lengths := []int{16, 16, 24, 24}
		tabsPrint(columns{"BALANCE", "PENDING CHARGES", "LAST PAYMENT DATE", "LAST PAYMENT AMOUNT"}, lengths)
		tabsPrint(columns{info.Balance, info.PendingCharges, info.LastPaymentDate, info.LastPaymentAmount}, lengths)
		tabsFlush()
	}
}
