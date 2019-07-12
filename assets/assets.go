package assets

import(

	"net/http"

	"github.com/rakyll/statik/fs"

	_ "github.com/HalalChain/qitmeer-wallet/assets/statik"
)

// GetStatic return statci binary fileSystem
func GetStatic()(http.FileSystem ,error){
	return fs.New()
}