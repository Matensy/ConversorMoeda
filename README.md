# 💱 Conversor de Moedas

Site que converte valores entre moedas e **criptomoedas** usando
**cotação em tempo real**. Você escolhe a moeda de origem, a de destino,
digita o valor e ele mostra quanto dá na outra moeda.

Dá para converter entre moedas tradicionais (Real, Dólar, Euro...),
entre criptos (Bitcoin, Ethereum...) e também misturar as duas
(ex: **0,5 BTC em Reais**). O cálculo usa o Dólar como "ponte".

## 🛠️ Tecnologias usadas

- **Back-end:** [Go (Golang)](https://go.dev/) — usando só a biblioteca padrão
- **Front-end:** HTML + [Bootstrap 5](https://getbootstrap.com/) + JavaScript
- **APIs (gratuitas e sem chave):**
  - [open.er-api.com](https://www.exchangerate-api.com/docs/free) — moedas tradicionais
  - [CoinGecko](https://www.coingecko.com/pt/api) — preço das criptomoedas

## 📁 Estrutura dos arquivos

```
ConversorMoeda/
├─ main.go              → o back-end (servidor + conversão)
├─ go.mod
└─ static/              → o site
   ├─ index.html
   ├─ css/style.css
   └─ js/app.js
```

## ▶️ Como rodar o projeto

> Antes de começar, você precisa ter o [Go](https://go.dev/dl/) instalado.

1. Abra o terminal dentro da pasta do projeto.

2. Rode o servidor:

   ```bash
   go run .
   ```

3. Abra no navegador:

   ```
   http://localhost:3000
   ```

## 🧭 Como funciona (passo a passo)

1. Você digita o valor e escolhe as moedas (de / para).
2. O site chama o back-end em `/api/converter`.
3. O back-end busca a cotação atual na API open.er-api.com e calcula o valor.
4. O resultado aparece formatado, junto com a taxa usada e a data da cotação.

O botão **⇄** inverte rapidamente a moeda de origem e a de destino.

## 💡 Observações

- A lista de moedas fica no `main.go` (variável `moedas`), fácil de adicionar
  mais.
- A conversão passa pelo back-end (e não direto do navegador), que é uma
  forma mais organizada de trabalhar com APIs.

---

Feito para estudo. 🚀
