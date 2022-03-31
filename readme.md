# dcard 2022 Backend Intern

 2022  縮網址dcard作業

## 製作心得

沒想到不知不覺就做出了V2版本來，原本是預設使用SQlite3來當作資料庫，但考量到這可能是需要快速回應的網站，所以改用redis，搭上rueidis套件。

由於先前都只是論壇上了解過redis，而沒有實作經驗，甚至不太懂redis本身也算是一個資料庫，在v1中，原本的理解是先透過 if看redis有沒有get到，如果有就return ，沒有就存進SQlite，然後在set redis一次，當然在做到一半的時候就打斷了這個做法。

## 主要工具

後端 -> gin

資料庫溝通套件 -> rueidis

爬蟲 -> colly

## 思考過程

### 如何判定是正確網址、有無效用

先使用正規化regex來判定字串是否符合，再採用colly爬蟲去判斷網站是否有回傳Statu 200，之後遇到有些伺服器會擋爬蟲，雖然有辦法可以閃避，但我覺得只要判定網址是否活著就行了。
之後改成使用net/http套件完成。

###  資料庫

採用redis，並直接搭配TTL倒數。

#### TTL設定

因為是第一次使用redis，所以我重新複習一次redis的用法，然後又跟go-redis比較後發現rueidis好像無法設定TTL，但我又想怎麼可能，所以我就慢慢除了README.md以外的文件，而果然真的有。

#### 計算相差時間

考量主機會架在台灣，但如果之後主機在非台灣地區的話時差會跑掉，所以再取得現在時間的地方加上設定時區。

取得現在時間後，將和用戶輸入的值進行比較，如果是0，照理來說應該要直接擋掉資料庫新增的方面會比較好，畢竟TTL為0是毫無意義的行為，但為了方便檢測這邊就沒有進行阻攔。

#### 如果相差範圍超過int64的話呢?

當初有思考過這一點，但後來測試的時候有發現golang會自動回傳變數的最大值，所以不會發生超出範圍的錯誤訊息。

## 踩坑

### colly 爬蟲

使用get方式且在帶參數時會出現如下錯誤訊息

目的網址:https://www.youtube.com/watch?v=rweQmBxc2lc&list=RDu0CqY27IFyo&index=4

    Could not open file '=RDu0CqY27IFyo': File not found'index' is not recognized as an internal or external command,operable program or batch file.

從 **=RDu0CqY27IFyo**的錯誤訊息回饋來看，應該是帶參數的問題

之後改用套件解決此問題。

## 進步方向

可以把router分開成獨立的go檔，且可以再根據不同API或是規模細分成多個go檔案來減輕單一go檔職責過重的問題。
