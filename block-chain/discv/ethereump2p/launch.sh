#!/bin/bash
xterm -e ./p2p-network -name node1 -port 30330 &
xterm -e ./p2p-network -name node2 -port 30331 -bootstrap enode://09d05fdfc30f024c75b4d7e304f9929616a1a525bcce90188fa5981596c732299715f3cef108ef7de9ee001643cafe2492b6036464050b5a35679f7cd55ccea7@127.0.0.1:30330 &
xterm -e ./p2p-network -name node3 -port 30332 -bootstrap enode://f2e908c249aa74eddaa1060c0cf067a713460c86f2cc99c952438756c60f7617a2bd9a3751e15edb5d39af3cd6da00752cf3eedaab5435fc35899afeda0d14d4@127.0.0.1:30331
