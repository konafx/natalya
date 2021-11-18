# 詠み人知らず
## データを読むタイミングを考える
- bot起動時にする
  - Pros
    - 一番良さそう、どうせ１台しか動かさないし
    - 割り切る
  - Cons
    - Canceled 対応がだるい。ちゃんとやらないと１日前から止まってたgameが、bot復活時に有効になってしまう
      - bulk update するのがよさそう。制限あるが、小規模と割り切るなら余裕だろう
- 各イベントで読む
  - Cons
    - これをやろうとしてだるいなーって言ってる
    - とにかくだるすぎる

## 操作
### 開始
- 起動
  - slash-command `/game poem start poet1:<member> poet2:<member>...`
  - [data] CREATE game
- DMで詩の一文字を送信
  - [data] game.poets

### 次
- gameが次のステージへ行ったら次のDMを促す
- gameのステージが 8とか7とか、とかく終わったらresult表示

### DM受信
- 歌詠み中のゲームから、DM相手が参加している場合、起動
  - [data] game by user
- 一文字判定
- 格納
  - [data] UPDATE game.poems by poet.NextPoetNumber

### 中断
- 起動
  - 開始時のメッセージにボタンつけときゃええか？
  - 離脱イベントで０人になったら
    - [data] game.NumberOfPoets
- 中断処理
  - [data] UPDATE game.status

### 離脱
- 起動
  - slash-command `/game poem leave`
  - [data] DELETE poet (and poem)
  - [data] UPDATE game.NumberOfPoets

## Domain
- game
  - status: enum
    - "composing", "composed", "canceled"
  - stage: number
  - numberOfPoets: number
  - []poet
    - id: string
    - nextPoemId
  - []poem
    - id
    - []poemRune
      - poetId: string
      - rune: string

poet には
- 参加しているゲーム
- 次どこを詠むか

が必要

### RDB的に

- game
  - guildId
  - channelId
  - has many poets

- poet
  - userId
  - gameId
    - has a game
  - nextPoemId
    - has a poem

- poem
  - has many poems

- poemRune
  - poemId
  - rune
  - poetId
    - has a poet
