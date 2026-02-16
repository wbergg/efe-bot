package sbfetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/efe-bot/config"
)

type SBAPIResponse struct {
	Metadata struct {
		DocCount               int `json:"docCount"`
		FullAssortmentDocCount int `json:"fullAssortmentDocCount"`
		NextPage               int `json:"nextPage"`
		PreviousPage           int `json:"previousPage"`
		TotalPages             int `json:"totalPages"`
		PriceRange             struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"priceRange"`
		VolumeRange struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"volumeRange"`
		AlcoholPercentageRange struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"alcoholPercentageRange"`
		SugarContentRange struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"sugarContentRange"`
		SugarContentGramPer100MlRange struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"sugarContentGramPer100mlRange"`
		DidYouMeanQuery interface{} `json:"didYouMeanQuery"`
	} `json:"metadata"`
	Products []struct {
		ProductID                 string        `json:"productId"`
		ProductNumber             string        `json:"productNumber"`
		ProductNameBold           string        `json:"productNameBold"`
		ProductNameThin           string        `json:"productNameThin"`
		Category                  interface{}   `json:"category"`
		ProductNumberShort        string        `json:"productNumberShort"`
		ProducerName              string        `json:"producerName"`
		SupplierName              string        `json:"supplierName"`
		IsKosher                  bool          `json:"isKosher"`
		BottleTextShort           string        `json:"bottleTextShort"`
		BottleText                string        `json:"bottleText"`
		RestrictedParcelQuantity  int           `json:"restrictedParcelQuantity"`
		IsOrganic                 bool          `json:"isOrganic"`
		IsSustainableChoice       bool          `json:"isSustainableChoice"`
		IsEthical                 bool          `json:"isEthical"`
		EthicalLabel              interface{}   `json:"ethicalLabel"`
		IsWebLaunch               bool          `json:"isWebLaunch"`
		ProductLaunchDate         string        `json:"productLaunchDate"`
		IsCompletelyOutOfStock    bool          `json:"isCompletelyOutOfStock"`
		IsTemporaryOutOfStock     bool          `json:"isTemporaryOutOfStock"`
		AlcoholPercentage         float64       `json:"alcoholPercentage"`
		Volume                    float64       `json:"volume"`
		Price                     float64       `json:"price"`
		Country                   string        `json:"country"`
		OriginLevel1              interface{}   `json:"originLevel1"`
		OriginLevel2              interface{}   `json:"originLevel2"`
		CategoryLevel1            string        `json:"categoryLevel1"`
		CategoryLevel2            string        `json:"categoryLevel2"`
		CategoryLevel3            string        `json:"categoryLevel3"`
		CategoryLevel4            interface{}   `json:"categoryLevel4"`
		CustomCategoryTitle       string        `json:"customCategoryTitle"`
		AssortmentText            string        `json:"assortmentText"`
		Usage                     string        `json:"usage"`
		Taste                     string        `json:"taste"`
		TasteSymbols              []string      `json:"tasteSymbols"`
		TasteClockGroupBitter     interface{}   `json:"tasteClockGroupBitter"`
		TasteClockGroupSmokiness  interface{}   `json:"tasteClockGroupSmokiness"`
		TasteClockBitter          int           `json:"tasteClockBitter"`
		TasteClockFruitacid       int           `json:"tasteClockFruitacid"`
		TasteClockBody            int           `json:"tasteClockBody"`
		TasteClockRoughness       int           `json:"tasteClockRoughness"`
		TasteClockSweetness       int           `json:"tasteClockSweetness"`
		TasteClockSmokiness       int           `json:"tasteClockSmokiness"`
		TasteClockCasque          int           `json:"tasteClockCasque"`
		Stock                     int           `json:"stock"`
		Shelf                     interface{}   `json:"shelf"`
		Assortment                string        `json:"assortment"`
		RecycleFee                float64       `json:"recycleFee"`
		IsManufacturingCountry    bool          `json:"isManufacturingCountry"`
		IsRegionalRestricted      bool          `json:"isRegionalRestricted"`
		IsInStoreSearchAssortment []interface{} `json:"isInStoreSearchAssortment"`
		Packaging                 string        `json:"packaging"`
		PackagingLevel1           string        `json:"packagingLevel1"`
		PackagingLevel2           interface{}   `json:"packagingLevel2"`
		PackagingCO2ImpactLevel   string        `json:"packagingCO2ImpactLevel"`
		PackagingTypeCode         string        `json:"packagingTypeCode"`
		IsNews                    bool          `json:"isNews"`
		Images                    []struct {
			ImageURL string      `json:"imageUrl"`
			FileType interface{} `json:"fileType"`
			Size     interface{} `json:"size"`
		} `json:"images"`
		IsDiscontinued                  bool          `json:"isDiscontinued"`
		IsSupplierTemporaryNotAvailable bool          `json:"isSupplierTemporaryNotAvailable"`
		SugarContent                    int           `json:"sugarContent"`
		SugarContentGramPer100Ml        float64       `json:"sugarContentGramPer100ml"`
		IsRecommendedByTasteProfile     interface{}   `json:"isRecommendedByTasteProfile"`
		Seal                            interface{}   `json:"seal"`
		Vintage                         []interface{} `json:"vintage"`
		OtherSelections                 interface{}   `json:"otherSelections"`
		TasteClocks                     []struct {
			Key   string `json:"key"`
			Value int    `json:"value"`
		} `json:"tasteClocks"`
		DishPoints         interface{} `json:"dishPoints"`
		NeedCrateProductID string      `json:"needCrateProductId"`
	} `json:"products"`
	SuggestedProducts []interface{} `json:"suggestedProducts"`
	Filters           []struct {
		Name                  string      `json:"name"`
		Type                  string      `json:"type"`
		DisplayName           string      `json:"displayName"`
		Description           string      `json:"description"`
		Summary               interface{} `json:"summary"`
		LegalText             interface{} `json:"legalText"`
		IsMultipleChoice      bool        `json:"isMultipleChoice"`
		IsActive              bool        `json:"isActive"`
		IsSubtitleTextVisible bool        `json:"isSubtitleTextVisible"`
		SearchModifiers       []struct {
			Value        string      `json:"value"`
			Count        int         `json:"count"`
			IsActive     bool        `json:"isActive"`
			SubtitleText interface{} `json:"subtitleText"`
		} `json:"searchModifiers"`
		Child struct {
			Name                  string      `json:"name"`
			Type                  string      `json:"type"`
			DisplayName           string      `json:"displayName"`
			Description           string      `json:"description"`
			Summary               interface{} `json:"summary"`
			LegalText             interface{} `json:"legalText"`
			IsMultipleChoice      bool        `json:"isMultipleChoice"`
			IsActive              bool        `json:"isActive"`
			IsSubtitleTextVisible bool        `json:"isSubtitleTextVisible"`
			SearchModifiers       []struct {
				Value        string      `json:"value"`
				Count        int         `json:"count"`
				IsActive     bool        `json:"isActive"`
				SubtitleText interface{} `json:"subtitleText"`
			} `json:"searchModifiers"`
			Child struct {
				Name                  string      `json:"name"`
				Type                  string      `json:"type"`
				DisplayName           string      `json:"displayName"`
				Description           interface{} `json:"description"`
				Summary               interface{} `json:"summary"`
				LegalText             interface{} `json:"legalText"`
				IsMultipleChoice      bool        `json:"isMultipleChoice"`
				IsActive              bool        `json:"isActive"`
				IsSubtitleTextVisible bool        `json:"isSubtitleTextVisible"`
				SearchModifiers       []struct {
					Value        string      `json:"value"`
					Count        int         `json:"count"`
					IsActive     bool        `json:"isActive"`
					SubtitleText interface{} `json:"subtitleText"`
				} `json:"searchModifiers"`
				Child struct {
					Name                  string        `json:"name"`
					Type                  string        `json:"type"`
					DisplayName           string        `json:"displayName"`
					Description           interface{}   `json:"description"`
					Summary               interface{}   `json:"summary"`
					LegalText             interface{}   `json:"legalText"`
					IsMultipleChoice      bool          `json:"isMultipleChoice"`
					IsActive              bool          `json:"isActive"`
					IsSubtitleTextVisible bool          `json:"isSubtitleTextVisible"`
					SearchModifiers       []interface{} `json:"searchModifiers"`
					Child                 interface{}   `json:"child"`
				} `json:"child"`
			} `json:"child"`
		} `json:"child"`
	} `json:"filters"`
	FilterMenuItems []interface{} `json:"filterMenuItems"`
}

type Result struct {
	NameBold string
	NameThin string
	Percent  float64
	Approved bool
}

func Get(config config.Config, search_string string) ([]Result, error) {

	// Search and url
	urlstr := config.SBAPI.Url

	// Build URL with query parameters
	search := url.Values{}
	search.Set("size", "30-50")
	search.Set("page", "1")
	search.Set("textQuery", search_string)

	fullUrl := fmt.Sprintf("%s?%s", urlstr, search.Encode())

	// Fetch
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		log.Error("Error creating request:", err)
		return []Result{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
	req.Header.Add("ocp-apim-subscription-key", config.SBAPI.Ocp_apim_subscription_key)
	req.Header.Set("Accept", "application/json")

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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("SBAPI returned status %d: %s", resp.StatusCode, string(body))
		return []Result{}, fmt.Errorf("SBAPI returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response:", err)
		return []Result{}, err
	}

	// Unmarshal
	var response SBAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshalling JSON:", err)
		return []Result{}, err
	}

	// Save to slice, only include beer products
	var results []Result
	for _, product := range response.Products {
		if !strings.EqualFold(product.CategoryLevel1, "Öl") {
			continue
		}
		result := Result{
			NameBold: product.ProductNameBold,
			NameThin: product.ProductNameThin,
			Percent:  product.AlcoholPercentage,
			Approved: product.AlcoholPercentage >= 5,
		}
		results = append(results, result)
	}

	return results, nil
}
