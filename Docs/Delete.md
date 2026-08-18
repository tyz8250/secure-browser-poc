# Delete Session

## do

### 2026-08-15

```
secure-browser-poc、今日は `DELETE /sessions/{id}` を実装。

ただ削除するだけじゃなくて、

- runningなら Stop → Remove
- exitedならそのまま Remove
- 存在しないIDは 404
- Docker操作失敗は 500
- 成功は 204 No Content

まで整理して実機確認した。

さらに「削除済みSessionをどう扱う？」まで考えると、Containerの存在だけじゃなく、誰がいつ削除したかをSession履歴として残したくなる。

コードを書く前に「正しい状態 → Failure Mode → 観測方法」を考えると、実装の見え方がだいぶ変わる。

```

memo
将来Session管理を永続化する場合は、未登録ID=404、削除済みSession=410 Goneとして区別し、deleted_by / deleted_atを監査履歴として保持する