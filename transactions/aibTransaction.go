package transactions

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/stonish/banking/utils"
)

type aibTransaction struct {
	id            string
	realDate      string
	completedDate time.Time
	details       string
	amount        float64
	balance       string
	source        string
	isHidden      bool
}

func NewAIBTransaction(rawData string) *aibTransaction {
	t := new(aibTransaction)

	data := strings.Split(rawData, ",")

	transactionDate, err := time.Parse("02/01/06", data[1])
	if err != nil {
		log.Fatal("Could not parse date: ", err, "\nRaw data: ", rawData)
	}

	t.completedDate = transactionDate

	t.details = strings.Trim(data[2], " ")
	t.details = strings.Replace(t.details, "\"", "", -1)

	t.amount, err = t.parseAmount(data[3], data[4])
	if err != nil {
		log.Fatal("Could not parse amount: ", err, "\nRaw data: ", rawData)
	}

	t.balance = data[5]
	t.source = "AIB"
	t.isHidden = t.shouldHide()
	t.id = utils.Classify(t.details)

	return t
}

func (t aibTransaction) GetCompletedDate() time.Time {
	return t.completedDate
}

func (t aibTransaction) Output() {
	if t.isHidden {
		return
	}

	roundedAmount := math.Round(t.amount*100) / 100
	strAmount := fmt.Sprintf("%f", roundedAmount)
	fmt.Println(t.id + "\t" +
		t.realDate + "\t" +
		t.completedDate.Format("2 Jan") + "\t" +
		strAmount + "\t" +
		t.details + "\t" +
		t.source)
}

func (t aibTransaction) parseAmount(debitAmount string, creditAmount string) (float64, error) {
	var amount float64

	if debitAmount != "" && creditAmount != "" {
		return amount, errors.New("Transaction with both credit and debit amounts found")
	}

	if debitAmount != "" {
		amount, err := strconv.ParseFloat(debitAmount, 32)

		if err != nil {
			return amount, fmt.Errorf("Could not parse amount: %w", err)
		}

		return amount, nil
	}

	if creditAmount != "" {
		amount, err := strconv.ParseFloat(creditAmount, 32)

		if err != nil {
			return amount, fmt.Errorf("Could not parse amount: %w", err)
		}

		if amount != 0 {
			amount = -1 * amount
		}

		return amount, nil
	}

	return amount, errors.New("Transaction with neither credit nor debit amounts found")
}

func (t aibTransaction) shouldHide() bool {
	if t.amount == 0 {
		return true
	}

	if t.isRevolutTopUp() {
		return true
	}

	if strings.HasPrefix(t.details, "*INET SAVINGS ") {
		return true
	}

	if t.amount == 600 &&
		(strings.HasPrefix(t.details, "*INET RENT ") ||
			strings.HasPrefix(t.details, "*INET DAVID ")) {
		return true
	}

	if t.amount == 10 && t.details == "931365 22689017" {
		return true
	}

	return false
}

func (t aibTransaction) isRevolutTopUp() bool {
	return strings.HasPrefix(t.details, "VDP-Revolut**") ||
		strings.HasPrefix(t.details, "VDP-Revolut* - ") ||
		strings.HasPrefix(t.details, "VDP-Revolut  - ") ||
		strings.HasPrefix(t.details, "VDP-REVOLUT*") ||
		t.details == "VDP-Revolut"
}
