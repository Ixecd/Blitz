## Bitcoin: A Peer-to-Peer Electronic Cash System

// tx (Transaction): UTXO spend/create, input sig your priv, output pubhash.
// 原子：in > out + fee。

// merkle_root: Binary tree of tx hashes, root in header。
// Proof: SPV light client O(log n) verify tx without full block。

// timestamp: ~10min, honest node reject >2h future/<local。

// nonce: 4byte counter, miner ++ grind hash(header) < target。

// difficulty: Target bits (e.g. 18 zero prefix), retarget 2016 blocks ~2w。
// Low difficulty = easy mine, high = hard。


Tx签名流程澄清（用Alice/Bob/Charlie，S5+ S8 reclaim）：

No上个私钥：每个input自己私钥签，验证自己pub（链式ownership）。
例子：

Genesis：Coinbase tx，miner create new coins to self pub。
Tx1 Alice→Bob：
Input: Alice's UTXO (e.g. coinbase)。
Alice hash(tx1), priv1 sign → sigScript。
Output: value to Bob pubhash。
Verify: Bob node check Alice pub verify sig。
Tx2 Bob→Charlie：
Input: Tx1 output (Bob's UTXO)。
Bob hash(tx2), priv2 sign → sigScript。
Output: value to Charlie pubhash。
Verify: Charlie node check Bob pub (from Tx1 output script) verify sig。

卧槽，我看懂了，就跟etcd之间peer用tls那个CA一模一样，只不过那个是环状的感觉，这个是链式的，每次输入都有节点自己的公钥，然后用自己的私钥签名发布给下一个接受者，之后下一个接受者根据上一个次的输入就知道上一个人的公钥，然后就可以验证签名这样子吗？

时间戳做hash块的时候，不仅要带着自己的时间戳，还要带着上一次交易的时间戳。