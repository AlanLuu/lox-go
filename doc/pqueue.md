# Priority queue methods

The following methods are defined in the built-in `pqueue` class:
- `pqueue class.args`
- `pqueue class.argsReversed`
- `pqueue class.builder()`
- `pqueue class.dict`
- `pqueue class.dictReversed`
- `pqueue class.dup()`
- `pqueue class.dupReversed()`
- `pqueue class.new()`
- `pqueue class.newReversed()`
- `pqueue class.reversed()`, which is an alias for `pqueue class.newReversed`

Priority queues have the following methods associated with them:
- `pqueue.add`, which is an alias for `pqueue.enqueue`
- `pqueue.addArgSwitched`, which is an alias for `pqueue.enqueueArgSwitched`
- `pqueue.allowDupPriorities`
- `pqueue.allowsDupPriorities()`
- `pqueue.clear()`
- `pqueue.contains`
- `pqueue.containsArgSwitched`
- `pqueue.containsPriority`
- `pqueue.containsValue`

- `pqueue.dequeue()`
- `pqueue.dequeueErr()`
- `pqueue.dequeueIter()`
- `pqueue.dequeuePriority()`
- `pqueue.dequeuePriorityErr()`
- `pqueue.dequeuePriorityIter()`
- `pqueue.dequeueValue()`
- `pqueue.dequeueValueErr()`
- `pqueue.dequeueValueIter()`
- `pqueue.enqueue`
- `pqueue.enqueueArgSwitched`
- `pqueue.equals`
- `pqueue.equalsPriorities`
- `pqueue.equalsValues`
- `pqueue.forEach`
- `pqueue.getPriorityByValue`
- `pqueue.getPriorityByValueErr`
- `pqueue.getValueByPriority`
- `pqueue.getValueByPriorityErr`
- `pqueue.isEmpty()`
- `pqueue.isReversed()`
- `pqueue.peek()`
- `pqueue.peekErr()`
- `pqueue.peekPriority()`
- `pqueue.peekPriorityErr()`
- `pqueue.peekValue()`
- `pqueue.peekValueErr()`
- `pqueue.print()`
- `pqueue.priorities()`
- `pqueue.prioritiesIter()`
- `pqueue.prioritiesList()`
- `pqueue.remove()`, which is an alias for `pqueue.dequeue`
- `pqueue.removeErr()`, which is an alias for `pqueue.dequeueErr`
- `pqueue.removeIter()`, which is an alias for `pqueue.dequeueIter`
- `pqueue.removePriority()`, which is an alias for `pqueue.dequeuePriority`
- `pqueue.removePriorityErr()`, which is an alias for `pqueue.dequeuePriorityErr`
- `pqueue.removePriorityIter()`, which is an alias for `pqueue.dequeuePriorityIter`
- `pqueue.removeValue()`, which is an alias for `pqueue.dequeueValue`
- `pqueue.removeValueErr()`, which is an alias for `pqueue.dequeueValueErr`
- `pqueue.removeValueIter()`, which is an alias for `pqueue.dequeueValueIter`
- `pqueue.reset()`
- `pqueue.resetReversed()`
- `pqueue.str()`
- `pqueue.toDict()`
- `pqueue.toList()`
- `pqueue.values()`
- `pqueue.valuesIter()`

Priority queue builders have the following methods associated with them:
- `pqueue builder.allowDupPriorities`
- `pqueue builder.build()`
- `pqueue builder.buildArgs`
- `pqueue builder.buildDict`
- `pqueue builder.isReversed`
- `pqueue builder.toDict()`
