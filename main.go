package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const porta = "3000"

// Moeda representa uma moeda (fiat) ou criptomoeda disponível.
type Moeda struct {
	Codigo string `json:"codigo"`
	Nome   string `json:"nome"`
	Tipo   string `json:"tipo"`   // "fiat" ou "cripto"
	CoinID string `json:"-"`      // id usado na API CoinGecko (só cripto)
}

// lista de moedas que aparecem no site.
var moedas = []Moeda{
	// Moedas tradicionais (fiat)
	{"BRL", "Real brasileiro", "fiat", ""},
	{"USD", "Dólar americano", "fiat", ""},
	{"EUR", "Euro", "fiat", ""},
	{"GBP", "Libra esterlina", "fiat", ""},
	{"JPY", "Iene japonês", "fiat", ""},
	{"ARS", "Peso argentino", "fiat", ""},
	{"CAD", "Dólar canadense", "fiat", ""},
	{"AUD", "Dólar australiano", "fiat", ""},
	{"CHF", "Franco suíço", "fiat", ""},
	{"CNY", "Yuan chinês", "fiat", ""},
	{"MXN", "Peso mexicano", "fiat", ""},
	{"UYU", "Peso uruguaio", "fiat", ""},
	// Criptomoedas
	{"BTC", "Bitcoin", "cripto", "bitcoin"},
	{"ETH", "Ethereum", "cripto", "ethereum"},
	{"USDT", "Tether", "cripto", "tether"},
	{"BNB", "BNB", "cripto", "binancecoin"},
	{"SOL", "Solana", "cripto", "solana"},
	{"XRP", "XRP", "cripto", "ripple"},
	{"ADA", "Cardano", "cripto", "cardano"},
	{"DOGE", "Dogecoin", "cripto", "dogecoin"},
	{"LTC", "Litecoin", "cripto", "litecoin"},
}

// cliente HTTP com tempo limite, para não travar se a API demorar.
var cliente = &http.Client{Timeout: 10 * time.Second}

// acha uma moeda da lista pelo código (ex: "BTC").
func acharMoeda(codigo string) (Moeda, bool) {
	for _, m := range moedas {
		if m.Codigo == codigo {
			return m, true
		}
	}
	return Moeda{}, false
}

// busca as cotações das moedas fiat (quantas de cada moeda vale 1 USD).
func buscarFiat() (map[string]float64, string, error) {
	resp, err := cliente.Get("https://open.er-api.com/v6/latest/USD")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var dados struct {
		Result     string             `json:"result"`
		Atualizado string             `json:"time_last_update_utc"`
		Rates      map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		return nil, "", err
	}
	if dados.Result != "success" {
		return nil, "", fmt.Errorf("falha na cotação fiat")
	}
	return dados.Rates, dados.Atualizado, nil
}

// busca o preço das criptomoedas em dólar (USD por 1 unidade da cripto).
func buscarCripto() (map[string]float64, error) {
	ids := []string{}
	for _, m := range moedas {
		if m.Tipo == "cripto" {
			ids = append(ids, m.CoinID)
		}
	}
	url := "https://api.coingecko.com/api/v3/simple/price?ids=" +
		strings.Join(ids, ",") + "&vs_currencies=usd"

	resp, err := cliente.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bruto map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&bruto); err != nil {
		return nil, err
	}

	precos := map[string]float64{}
	for coinID, valores := range bruto {
		precos[coinID] = valores["usd"]
	}
	return precos, nil
}

// calcula quantos dólares vale 1 unidade da moeda.
func usdPorUnidade(m Moeda, fiat map[string]float64, cripto map[string]float64) (float64, bool) {
	if m.Tipo == "cripto" {
		v, ok := cripto[m.CoinID]
		return v, ok && v > 0
	}
	r, ok := fiat[m.Codigo]
	if !ok || r == 0 {
		return 0, false
	}
	return 1.0 / r, true // se 1 USD = r da moeda, então 1 da moeda = 1/r USD
}

// GET /api/moedas -> lista as moedas disponíveis.
func handlerMoedas(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, moedas)
}

// GET /api/converter?de=BTC&para=BRL&valor=0.5
func handlerConverter(w http.ResponseWriter, r *http.Request) {
	deCod := strings.ToUpper(r.URL.Query().Get("de"))
	paraCod := strings.ToUpper(r.URL.Query().Get("para"))
	valorTexto := r.URL.Query().Get("valor")

	if deCod == "" || paraCod == "" || valorTexto == "" {
		responderErro(w, "Informe a moeda de origem, destino e o valor.", http.StatusBadRequest)
		return
	}

	valor, err := strconv.ParseFloat(strings.Replace(valorTexto, ",", ".", 1), 64)
	if err != nil || valor <= 0 {
		responderErro(w, "Valor inválido.", http.StatusBadRequest)
		return
	}

	de, ok1 := acharMoeda(deCod)
	para, ok2 := acharMoeda(paraCod)
	if !ok1 || !ok2 {
		responderErro(w, "Moeda inválida.", http.StatusBadRequest)
		return
	}

	// sempre precisamos das cotações fiat (o USD é a "ponte").
	fiat, atualizado, err := buscarFiat()
	if err != nil {
		responderErro(w, "Não foi possível consultar as cotações agora.", http.StatusBadGateway)
		return
	}

	// só busca criptos se alguma das moedas for cripto.
	var cripto map[string]float64
	if de.Tipo == "cripto" || para.Tipo == "cripto" {
		cripto, err = buscarCripto()
		if err != nil {
			responderErro(w, "Não foi possível consultar o preço das criptos agora.", http.StatusBadGateway)
			return
		}
	}

	usdDe, ok1 := usdPorUnidade(de, fiat, cripto)
	usdPara, ok2 := usdPorUnidade(para, fiat, cripto)
	if !ok1 || !ok2 || usdPara == 0 {
		responderErro(w, "Não foi possível obter a cotação dessas moedas.", http.StatusBadGateway)
		return
	}

	taxa := usdDe / usdPara // 1 "de" = taxa "para"

	responderJSON(w, map[string]any{
		"de":         de.Codigo,
		"para":       para.Codigo,
		"valor":      valor,
		"taxa":       taxa,
		"resultado":  valor * taxa,
		"atualizado": atualizado,
	})
}

func responderJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func responderErro(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"erro": msg})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	mux.Handle("/css/", http.FileServer(http.Dir("static")))
	mux.Handle("/js/", http.FileServer(http.Dir("static")))

	mux.HandleFunc("/api/moedas", handlerMoedas)
	mux.HandleFunc("/api/converter", handlerConverter)

	fmt.Printf("Servidor rodando! Abra no navegador: http://localhost:%s\n", porta)
	log.Fatal(http.ListenAndServe(":"+porta, mux))
}
