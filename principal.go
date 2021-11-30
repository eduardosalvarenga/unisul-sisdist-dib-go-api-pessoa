package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type Pessoa struct {
	Id 				int 		`json:"id,omitempty"`
	PrimeiroNome 	string 		`json:"nome,omitempty"`
	Sobrenome 		string 		`json:"sobrenome,omitempty"`
	Endereco 		*Endereco 	`json:"enderecos,omitempty"`
}

type Endereco struct {
	Id 			int 	`json:"endereco_id,omitempty"`
	Logradouro string `json:"logradouro,omitempty"`
	Cep        int    `json:"cep,omitempty"`
	Bairro     string `json:"bairro,omitempty"`
	Cidade 		string 	`json:"cidade,omitempty"`
	UF 			string 	`json:"uf,omitempty"`
}

type Cidade struct {
	Cidade 		string 	`json:"cidade,omitempty"`
	UF 			string 	`json:"uf,omitempty"`
	Pessoas		[]Pessoa `json:"pessoas,omitempty"`
}

const (
	user = "postgres"
	password ="postgres"
	baseDados = "sist_distrib_fp"
	host="localhost"
	port="6543"
)

func conectaNoBancoDeDados() *sql.DB {
	conexao := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", user, baseDados, password, host, port)
	db, err := sql.Open("postgres", conexao)
	mensagemErro(err)
	return db
}

func mensagemErro(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func ListarPessoas(w http.ResponseWriter, r *http.Request) {
	db := conectaNoBancoDeDados()
	listaPessoas, err := db.Query("select p.id, p.nome, p.sobrenome, p.endereco_id, e.logradouro, e.Cep, e.bairro, e.cidade, e.uf from tb_pessoa p left join tb_endereco e on p.endereco_id=e.id")
	mensagemErro(err)
	var pessoas []Pessoa
	for listaPessoas.Next() {
		var (
			id, endereco_id, cep int
			nome, sobrenome, logradouro, bairro, cidade, uf string
		)
		err = listaPessoas.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &cep, &bairro, &cidade, &uf)
		mensagemErro(err)
		pessoa := Pessoa{}
		pessoa.Id=id
		pessoa.PrimeiroNome=nome
		pessoa.Sobrenome=sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		pessoa.Endereco= &endereco
		pessoas = append(pessoas, pessoa)
	}
	defer db.Close()
	//fmt.Println(pessoas)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoas)

}

func BuscarPessoaPorID(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	idPessoa := parametros["ID"]
	pessoa := consultarPessoa(idPessoa)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoa)

}

func consultarPessoa(idPessoa string) Pessoa {
	db := conectaNoBancoDeDados()
	selecaoPorID, err := db.Query("select p.id, p.nome, p.sobrenome, p.endereco_id, e.logradouro, e.Cep, e.bairro, e.cidade, e.uf from tb_pessoa p left join tb_endereco e on p.endereco_id=e.id where p.id=$1", idPessoa)
	mensagemErro(err)
	pessoa := Pessoa{}
	for selecaoPorID.Next() {
		var (
			id, endereco_id, cep                            int
			nome, sobrenome, logradouro, bairro, cidade, uf string
		)
		err = selecaoPorID.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &cep, &bairro, &cidade, &uf)
		mensagemErro(err)

		pessoa.Id = id
		pessoa.PrimeiroNome = nome
		pessoa.Sobrenome = sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		pessoa.Endereco = &endereco
	}
	defer db.Close()
	return pessoa
}

func BuscarPessoaPorNomeESobrenome(w http.ResponseWriter, r *http.Request) {
	nomeURL := r.URL.Query().Get("nome")
	sobrenomeURL := r.URL.Query().Get("sobrenome")
	db := conectaNoBancoDeDados()
	selecaoPorID, err := db.Query("select p.id, p.nome, p.sobrenome, p.endereco_id, e.logradouro, e.Cep, e.bairro, e.cidade, e.uf from tb_pessoa p left join tb_endereco e on p.endereco_id=e.id where p.nome=$1 and p.sobrenome=$2", nomeURL, sobrenomeURL)
	mensagemErro(err)
	pessoa := Pessoa{}
	for selecaoPorID.Next() {
		var (
			id, endereco_id, cep int
			nome, sobrenome, logradouro, bairro, cidade, uf string
		)
		err = selecaoPorID.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &cep, &bairro, &cidade, &uf)
		mensagemErro(err)

		pessoa.Id=id
		pessoa.PrimeiroNome=nome
		pessoa.Sobrenome=sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		pessoa.Endereco= &endereco
	}
	defer db.Close()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoa)
}

func CriarPessoa(w http.ResponseWriter, r *http.Request) {
	var pessoa Pessoa
	json.NewDecoder(r.Body).Decode(&pessoa)
	db := conectaNoBancoDeDados()
	inserindoDados, err := db.Prepare("insert into tb_pessoa (id, nome, sobrenome, endereco_id)	values (nextval('seq_pessoa'),$1, $2, $3)")
	mensagemErro(err)
	inserindoDados.Exec(pessoa.PrimeiroNome, pessoa.Sobrenome,pessoa.Endereco.Id)
	defer db.Close()
}

func AlterarPessoa(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	var pessoaNova Pessoa
	err := json.NewDecoder(r.Body).Decode(&pessoaNova)
	mensagemErro(err)
	db := conectaNoBancoDeDados()
	alterandoPessoa, err := db.Prepare("update tb_pessoa set nome=$1, sobrenome=$2, endereco_id=$3 where id=$4")
	mensagemErro(err)
	alterandoPessoa.Exec(pessoaNova.PrimeiroNome, pessoaNova.Sobrenome,pessoaNova.Endereco.Id, parametros["ID"])
	defer db.Close()
	pessoa := consultarPessoa(parametros["ID"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoa)

}

func DeletarPessoa(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	db := conectaNoBancoDeDados()
	deletandoPessoa, err := db.Prepare("delete from tb_pessoa where id=$1")
	mensagemErro(err)
	deletandoPessoa.Exec(parametros["ID"])
	defer db.Close()
}

func BuscarPessoasPorCidade(w http.ResponseWriter, r *http.Request){
	cidadeURL := r.URL.Query().Get("cidade")
	db := conectaNoBancoDeDados()
	selecaoPorID, err := db.Query("select p.id, p.nome, p.sobrenome, e.cidade, e.uf from tb_pessoa p left join tb_endereco e on p.endereco_id=e.id where e.cidade=$1", cidadeURL)
	mensagemErro(err)
	pessoasPorCidade := Cidade{}
	var pessoas []Pessoa
	for selecaoPorID.Next() {
		var (
			id int
			nome, sobrenome, cidade, uf string
		)
		err = selecaoPorID.Scan(&id, &nome, &sobrenome, &cidade, &uf)
		mensagemErro(err)
		pessoasPorCidade.Cidade=cidade
		pessoasPorCidade.UF=uf
		pessoa := Pessoa{Id: id, PrimeiroNome: nome, Sobrenome: sobrenome}
		pessoas = append(pessoas, pessoa)
		pessoasPorCidade.Pessoas= pessoas
	}
	defer db.Close()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pessoasPorCidade)
}

func main() {
	rotas := mux.NewRouter()
	rotas.HandleFunc("/pessoas", ListarPessoas).Methods("GET")
	rotas.HandleFunc("/pessoa/{ID}", BuscarPessoaPorID).Methods("GET")
	rotas.HandleFunc("/pessoa/nome/", BuscarPessoaPorNomeESobrenome).Methods("GET")
	rotas.HandleFunc("/pessoas/cidade/", BuscarPessoasPorCidade).Methods("GET")
	rotas.HandleFunc("/pessoa/", CriarPessoa).Methods("POST")
	rotas.HandleFunc("/pessoa/{ID}", AlterarPessoa).Methods("PUT")
	rotas.HandleFunc("/pessoa/{ID}", DeletarPessoa).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":9090",  rotas))
}
