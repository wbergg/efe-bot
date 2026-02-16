package bsfetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/efe-bot/config"
)

type BSAPIResponse struct {
	Products []struct {
		IsCheapest       bool `json:"isCheapest"`
		LeftBottomSplash bool `json:"leftBottomSplash"`
		Discount         struct {
			MaxDiscountedItems int64 `json:"maxDiscountedItems"`
			SingleUnitPrice    struct {
				AmountAsDecimal float64 `json:"amountAsDecimal"`
				Amount          string  `json:"amount"`
				Major           string  `json:"major"`
				Minor           string  `json:"minor"`
			} `json:"singleUnitPrice"`
			NumberOfItemsNeeded int    `json:"numberOfItemsNeeded"`
			ShowPriceForOne     bool   `json:"showPriceForOne"`
			IsSmileOffer        bool   `json:"isSmileOffer"`
			DiscountText        string `json:"discountText"`
			BeforePrice         struct {
				AmountAsDecimal float64 `json:"amountAsDecimal"`
				Amount          string  `json:"amount"`
				Major           string  `json:"major"`
				Minor           string  `json:"minor"`
			} `json:"beforePrice"`
			SplashText        string `json:"splashText"`
			BeforePricePrefix string `json:"beforePricePrefix"`
		} `json:"discount"`
		LeftSplash struct {
			Type int `json:"type"`
		} `json:"leftSplash"`
		Uom                  string `json:"uom"`
		QtyPrUom             string `json:"qtyPrUom"`
		UnitPriceText1       string `json:"unitPriceText1"`
		UnitPriceText2       string `json:"unitPriceText2"`
		ID                   string `json:"id"`
		ProductClickTracking string `json:"productClickTracking"`
		AddToBasket          struct {
			DisplayName         string `json:"displayName"`
			PrimaryCategory     string `json:"primaryCategory"`
			MinimumQuantityText string `json:"minimumQuantityText"`
			MinimumQuantity     int    `json:"minimumQuantity"`
			ID                  string `json:"id"`
			Ean                 string `json:"ean"`
			InitialQuantity     int    `json:"initialQuantity"`
			IsShopOnly          bool   `json:"isShopOnly"`
			IsSoldOut           bool   `json:"isSoldOut"`
			ProductID           string `json:"productId"`
		} `json:"addToBasket"`
		Price struct {
			AmountAsDecimal float64 `json:"amountAsDecimal"`
			Amount          string  `json:"amount"`
			Major           string  `json:"major"`
			Minor           string  `json:"minor"`
		} `json:"price"`
		DisplayName string `json:"displayName"`
		Image       string `json:"image"`
		URL         string `json:"url"`
		Brand       string `json:"brand,omitempty"`
	} `json:"products"`
	Facets          []interface{} `json:"facets"`
	ChildCategories []interface{} `json:"childCategories"`
	Total           int           `json:"total"`
	IsEmpty         bool          `json:"isEmpty"`
}

var percentRegex = regexp.MustCompile(`\s([0-9]+(?:[,.][0-9]+)?)\s*%`)

type Result struct {
	NameBold string
	NameThin string
	Percent  float64
	Approved bool
}

func Get(config config.Config, search_string string) ([]Result, error) {

	// Build URL - config URL already includes ?pageSize=100&term=
	fullUrl := config.BSAPI.Url + url.QueryEscape(search_string)

	// Fetch
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		log.Error("Error creating request:", err)
		return []Result{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")

	// Client shit
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request:", err)
		return []Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response:", err)
		return []Result{}, err
	}

	// Unmarshal
	var response BSAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshalling JSON:", err)
		return []Result{}, err
	}

	// Save to slice
	var results []Result
	for _, product := range response.Products {
		percent, err := GetPercent(product.DisplayName)
		if err != nil {
			// Log the error and skip this product
			log.Warnf("Skipping product due to parse error: %v", err)
			continue
		}
		result := Result{
			NameBold: product.DisplayName,
			NameThin: "",
			Percent:  percent,
			Approved: percent >= 5,
		}
		results = append(results, result)
	}

	return results, nil
}

func GetPercent(input string) (float64, error) {
	match := percentRegex.FindStringSubmatch(input)

	// If no match found, return error
	if match == nil {
		return 0, fmt.Errorf("no alcohol percentage found in product name: %s", input)
	}

	s := strings.Replace(match[1], ",", ".", 1)

	percent, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse alcohol percentage '%s': %w", s, err)
	}

	return percent, nil
}
