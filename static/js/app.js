const selDe = document.getElementById("de");
const selPara = document.getElementById("para");
const campoValor = document.getElementById("valor");
const form = document.getElementById("form");
const divErro = document.getElementById("erro");
const divResultado = document.getElementById("resultado");

// guarda o tipo de cada moeda (fiat ou cripto) para saber como formatar.
const tipoPorCodigo = {};

function formatar(valor, moeda) {
  // criptomoedas não têm formato de moeda padrão: mostramos com mais casas.
  if (tipoPorCodigo[moeda] === "cripto") {
    const num = valor.toLocaleString("pt-BR", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 8,
    });
    return num + " " + moeda;
  }
  try {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: moeda,
    }).format(valor);
  } catch (e) {
    return valor.toFixed(2) + " " + moeda;
  }
}

async function carregarMoedas() {
  const resp = await fetch("/api/moedas");
  const lista = await resp.json();

  // monta os dois selects de forma independente.
  lista.forEach((m) => (tipoPorCodigo[m.codigo] = m.tipo));
  preencher(selDe, lista);
  preencher(selPara, lista);

  selDe.value = "BTC";
  selPara.value = "BRL";
}

function preencher(sel, lista) {
  const grupoFiat = document.createElement("optgroup");
  grupoFiat.label = "Moedas";
  const grupoCripto = document.createElement("optgroup");
  grupoCripto.label = "Criptomoedas";

  lista.forEach((m) => {
    const op = new Option(m.codigo + " - " + m.nome, m.codigo);
    (m.tipo === "cripto" ? grupoCripto : grupoFiat).appendChild(op);
  });

  sel.appendChild(grupoFiat);
  sel.appendChild(grupoCripto);
}

document.getElementById("inverter").addEventListener("click", () => {
  const tmp = selDe.value;
  selDe.value = selPara.value;
  selPara.value = tmp;
});

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  divErro.classList.add("d-none");
  divResultado.classList.add("d-none");

  const params = new URLSearchParams({
    de: selDe.value,
    para: selPara.value,
    valor: campoValor.value,
  });

  const resp = await fetch("/api/converter?" + params.toString());
  const dados = await resp.json();

  if (!resp.ok) {
    divErro.textContent = dados.erro || "Não foi possível converter.";
    divErro.classList.remove("d-none");
    return;
  }

  document.getElementById("origem").textContent =
    formatar(dados.valor, dados.de) + " equivale a";
  document.getElementById("valorConvertido").textContent = formatar(
    dados.resultado,
    dados.para
  );
  document.getElementById("taxa").textContent =
    "1 " + dados.de + " = " + formatar(dados.taxa, dados.para);
  document.getElementById("atualizado").textContent =
    "Atualizado: " + dados.atualizado;
  divResultado.classList.remove("d-none");
});

carregarMoedas();
