/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/19
   Description :
-------------------------------------------------
*/

package utils

var Text = &textUtil{}

type textUtil struct{}

// 模糊匹配, ? 表示一个字符, * 表示任意字符串或空字符串
func (*textUtil) IsMatchWildcard(text string, p string) bool {
	m, n := len(text), len(p)
	dp := make([][]bool, m+1)
	for i := 0; i <= m; i++ {
		dp[i] = make([]bool, n+1)
	}
	dp[0][0] = true
	for i := 1; i <= n; i++ {
		if p[i-1] == '*' {
			dp[0][i] = true
		} else {
			break
		}
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if p[j-1] == '*' {
				dp[i][j] = dp[i][j-1] || dp[i-1][j]
			} else if p[j-1] == '?' || text[i-1] == p[j-1] {
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}
	return dp[m][n]
}

// 模糊匹配, 同 IsMatchWildcard, 只要匹配某一个通配符则返回true
func (u *textUtil) IsMatchWildcardAny(text string, ps ...string) bool {
	for _, p := range ps {
		if u.IsMatchWildcard(text, p) {
			return true
		}
	}
	return false
}
