## Description ##
This is an utility for [The Mathless Kungfu](https://store.steampowered.com/app/1696440/The_Matchless_Kungfu/) game.
It looks for an optimal (smallest) string that combines multiple Inner Kunfu Techniques.
Each one is represented by a string of symbols of A for triangle, N for square and O for circle chi nodes.
The choice of characters is arbitrary and you can use others, as long as they are consistent with other definitions in 
combination.yaml
You can use another file with combinations using "-filename other_file.yaml"  command line parameter

## Installation instructions ##
1. Install go and git
2. Checkout this repo with "git clone git@github.com:konstmonst/matchless_kungfu_combo_creator.git" 
3. Run "go mod tidy" in project's root directory
4. Run "go build"
5. Run resulting executable and edit combinations.xml for combinations
