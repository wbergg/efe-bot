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

type Result struct {
	NameBold string
	NameThin string
	Percent  float64
	Approved bool
}

func Get(cfg string, search_string string) ([]Result, error) {

	// Load config
	config, err := config.LoadConfig(cfg)
	if err != nil {
		return []Result{}, fmt.Errorf("could not load config: %w", err)
	}

	// Url
	urlstr := config.BSAPI.Url

	// Build URL with query parameters
	search := url.Values{}
	search.Set("term", search_string)

	fullUrl := fmt.Sprintf("%s%s", urlstr, search.Encode())
	fmt.Println(fullUrl)

	// Fetch
	req, err := http.NewRequest("GET", fullUrl, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")

	if err != nil {
		log.Error("Error fetching data:", err)
		return []Result{}, err
	}

	// Client shit
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request:", err)
		return []Result{}, err
	}
	fmt.Println(resp)
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
	re := regexp.MustCompile(`\s[0-9],[0-9]%`)
	match := re.FindString(input)

	// If no match found, return error
	if match == "" {
		return 0, fmt.Errorf("no alcohol percentage found in product name: %s", input)
	}

	s := strings.Replace(match, ",", ".", 1)
	s2 := strings.Replace(s, "%", "", 1)
	s3 := strings.Replace(s2, " ", "", 1)

	s4, err := strconv.ParseFloat(s3, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse alcohol percentage '%s': %w", s3, err)
	}

	return s4, nil
}
