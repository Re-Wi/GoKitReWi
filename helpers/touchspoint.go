package helpers

// type MyJson struct {
// 	MySample []struct {
// 		data map[string]string
// 	}
// }

//	type PublicKey struct {
//		Name  string
//		Price string
//		List  string
//	}

// type PublicKey struct {
// 	Name  string `json:"name"`
// 	Price string `json:"price"`
// 	List  string `json:"list"`
// }

type PublicKey struct {
	Id         int    `json:"id"`
	Signal     int    `json:"signal"`
	Count      int    `json:"count"`
	Points     string `json:"points"`
	CreateTime string `json:"create_time"`
}

type KeysResponse struct {
	Collection []PublicKey
}

var InfoLine = []interface{}{nil, nil, nil, nil, nil, nil}
