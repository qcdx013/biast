package main

import (
	"net/http"
	"sync"
)

var author2Articles = map[string][]aid{} // []aid 's ascending orderd(instead of descending orderd, think why~)
var authorMutex sync.RWMutex

func authorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	author := r.URL.Path[len(config["AuthorUrl"]):]
	if indexList := getArticleByAuthor(author); indexList == nil {
		http.NotFound(w, r)
	} else {
		if err := tmpl.ExecuteTemplate(w, "index", map[string]interface{}{
			"config":         config,
			"articles":       indexList,
			"getCommentList": getCommentList,
			"header":         "Author: " + author,
		}); err != nil {
			logger.Println("author:", err.Error())
		}
	}
}

func getArticleByAuthor(author string) []*Article {
	var ret []*Article
	authorMutex.RLock()
	for _, id := range author2Articles[author] {
		ret = append(ret, getArticle(id))
	}
	authorMutex.RUnlock()
	return ret
}

func updateAuthor(id aid, old, author string) {
	println(id, old, author)
	if old == author {
		return
	}
	authorMutex.Lock()
	if old != "" {
		list := author2Articles[old]
		for i := len(list) - 1; i >= 0; i-- {
			if list[i] == id {
				for j := i; j < len(list)-1; j++ {
					list[j] = list[j+1]
				}
				list = list[:len(list)-1]
				break
			}
		}
	}
	author2Articles[author] = append(author2Articles[author], 0)
	list := author2Articles[author]
	for i := len(list) - 2; i >= 0; i-- {
		if list[i] < id {
			list[i+1] = list[i]
		} else {
			list[i+1] = id
			break
		}
	}
	authorMutex.Unlock()
}

func initPageAuthors() {
	config["AuthorUrl"] = config["RootUrl"] + "author/"
	http.HandleFunc(config["AuthorUrl"], getGzipHandler(authorHandler))
	articleList := getArticleList()
	sortSlice(articleList, func(a, b interface{}) bool {
		return a.(*Article).Id > b.(*Article).Id
	})
	for _, article := range articleList {
		author2Articles[article.Author] = append(author2Articles[article.Author], article.Id)
	}
}