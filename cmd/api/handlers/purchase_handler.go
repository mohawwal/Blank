package handlers

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"whatsapp-bot-app/internal/models"

// 	"github.com/labstack/echo/v4"
// )

// // ConfirmPurchase handles purchase confirmation via link click
// func (h *Handler) ConfirmPurchase(c echo.Context) error {
// 	txnIDStr := c.Param("txn_id")

// 	txnID, err := strconv.Atoi(txnIDStr)
// 	if err != nil {
// 		return c.HTML(http.StatusBadRequest, "<h1>Invalid transaction ID</h1>")
// 	}

// 	var txn models.TransactionModel
// 	if err := h.DB.Preload("User").First(&txn, txnID).Error; err != nil {
// 		log.Printf("Transaction not found: %v", err)
// 		return c.HTML(http.StatusNotFound, "<h1>Transaction not found</h1>")
// 	}

// 	var user models.UserModel
// 	if err := h.DB.First(&user, txn.UserID).Error; err != nil {
// 		log.Printf("User not found: %v", err)
// 		return c.HTML(http.StatusNotFound, "<h1>User not found</h1>")
// 	}

// 	// Check if already processed
// 	if txn.Status != "pending" {
// 		return c.HTML(http.StatusOK, fmt.Sprintf("<h1>Transaction already %s</h1>", txn.Status))
// 	}

// 	// Update status to confirmed
// 	txn.Status = "confirmed"
// 	h.DB.Save(&txn)

// 	// Notify user via WhatsApp
// 	h.TwilioService.SendMessage("whatsapp:"+user.PhoneNumber, "⏳ Processing your purchase... Please wait.")

// 	// Process in background
// 	go h.processPurchase(&txn, &user)

// 	return c.HTML(http.StatusOK, `
// 		<html>
// 		<head><title>Purchase Confirmed</title></head>
// 		<body style="font-family: Arial; text-align: center; padding: 50px;">
// 			<h1>Purchase Confirmed!</h1>
// 			<p>Your purchase is being processed.</p>
// 			<p>You'll receive a WhatsApp message shortly.</p>
// 			<p style="color: #666; margin-top: 30px;">You can close this page.</p>
// 		</body>
// 		</html>
// 	`)
// }

// func (h *Handler) processPurchase(txn *models.TransactionModel, user *models.UserModel) {
// 	// Update status
// 	txn.Status = "processing"
// 	h.DB.Save(txn)

// 	// 1. Charge the card via Paystack
// 	chargeResp, err := h.PaystackService.ChargeAuthorization(user.AuthorizationCode, user.Email, txn.Amount)
// 	if err != nil {
// 		log.Printf("Charge failed: %v", err)
// 		txn.Status = "failed"
// 		h.DB.Save(txn)
// 		h.TwilioService.SendMessage("whatsapp:"+user.PhoneNumber,
// 			fmt.Sprintf("Payment failed: %s\n\nPlease try again or contact support.", err.Error()))
// 		return
// 	}

// 	txn.PaystackRef = chargeResp.Data.Reference
// 	h.DB.Save(txn)

// 	// 2. Purchase airtime/data via VTPass
// 	requestID := fmt.Sprintf("TXN_%d_%s", txn.ID, txn.PaystackRef)
// 	txn.VTPassRequestID = requestID

// 	var vtpassErr error
// 	var productName string

// 	if txn.Type == "airtime" {
// 		resp, err := h.VTPassService.BuyAirtime(txn.Network, txn.Amount, txn.PhoneNumber, requestID)
// 		if err != nil {
// 			vtpassErr = err
// 		} else {
// 			txn.VTPassTxnID = resp.Data.TransactionID
// 			productName = resp.Data.ProductName
// 		}
// 	} else if txn.Type == "data" {
// 		resp, err := h.VTPassService.BuyData(txn.Network, txn.VariationCode, txn.PhoneNumber, requestID)
// 		if err != nil {
// 			vtpassErr = err
// 		} else {
// 			txn.VTPassTxnID = resp.Data.TransactionID
// 			productName = resp.Data.ProductName
// 		}
// 	}

// 	if vtpassErr != nil {
// 		log.Printf("VTPass purchase failed: %v", vtpassErr)
// 		txn.Status = "failed"
// 		h.DB.Save(txn)
// 		h.TwilioService.SendMessage("whatsapp:"+user.PhoneNumber,
// 			"Purchase failed after payment was charged.\n\n"+
// 			"Your money will be refunded within 24 hours.\n\n"+
// 			"Contact support if you need assistance.")
// 		return
// 	}

// 	// 3. Success!
// 	txn.Status = "completed"
// 	h.DB.Save(txn)

// 	msg := fmt.Sprintf("✅ Purchase Successful!\n\n"+
// 		"📦 Product: %s\n"+
// 		"📱 Number: %s\n"+
// 		"💰 Amount: ₦%d\n"+
// 		"🔖 Reference: %s\n\n"+
// 		"Thank you for using Blank AI!",
// 		productName, txn.PhoneNumber, txn.Amount, txn.VTPassTxnID)

// 	h.TwilioService.SendMessage("whatsapp:"+user.PhoneNumber, msg)
// }
