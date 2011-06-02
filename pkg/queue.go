package orchestra;

/* Generic Concurrent-safe Queue w/ R/W locking, because I keep on 
 * reimplementing this.
*/

import "sync"

type QueueIter struct {
	cur	*QueueItem
	q	*Queue
}

/* Get the next item */
func (iter *QueueIter) Next() (item interface{}) {
	iter.q.l.RLock()
	if (nil == iter.cur) {
		item = nil
	} else {
		item = iter.cur.item
		iter.cur = iter.cur.next
	}
	iter.q.l.RUnlock()
	return item
}

/* Reset the iterator */
func (iter *QueueIter) Reset() {
	iter.q.l.RLock()
	iter.cur = iter.q.head
	iter.q.l.RUnlock()
}

type QueueItem struct {
	item	interface{}
	next	*QueueItem
}

type Queue struct {
	head	*QueueItem
	tail	*QueueItem
	l	sync.RWMutex
}

/* Initialise a new queue */
func NewQueue() (q *Queue) {
	q = new(Queue)
	return q
}

/* Append to the tail */
func (q *Queue) Append(i interface{}) {
	qi := new(QueueItem)
	qi.item = i

	q.l.Lock()
	if (nil == q.tail) {
		q.head = qi
	} else {
		q.tail.next = qi
	}
	q.tail = qi
	qi.next = nil
	q.l.Unlock()
}

/* Insert to the head */
func (q *Queue) Insert(i interface{}) {
	qi := new(QueueItem)
	qi.item = i

	q.l.Lock()
	qi.next = q.head
	q.head = qi
	if q.head.next == nil {
		q.tail = q.head
	}
	q.l.Unlock()
}

/* Acquire item from the head */
func (q *Queue) Shift() (i interface{}) {
	var qi *QueueItem
	q.l.Lock()
	if (q.head == nil) {
		qi = nil
	} else {
		qi = q.head
		q.head = qi.next
		qi.next = nil
	}
	if (q.head == nil) {
		q.tail = nil
	}
	q.l.Unlock()

	if (qi == nil) {
		return nil
	}
	return qi.item
}

/* Acquire item from the tail */
func (q *Queue) Pop() (i interface{}) {
	var qi *QueueItem

	q.l.Lock()
	if (q.tail == nil) {
		qi = nil
	} else {
		qi = q.tail
		/* GOHACK! */
		if (qi == q.head) {
			q.head = nil
			q.tail = nil
		} else {
			for q.tail = q.head; q.tail != nil && q.tail.next != qi; q.tail = q.tail.next {}
			q.tail.next = nil
		}
		qi.next = nil
	}
	q.l.Unlock()

	if (qi == nil) {
		return nil
	}
	return qi.item
}

func (q *Queue) Remove(item interface{}) (found bool) {
	found = false
	q.l.Lock()
	if (q.head != nil) {
		if (q.head.item == item) {
			/* special case. */
			q.head = q.head.next
			found = true
		} else {
			/* find the preceding item */
			p := q.head
			for p.next != nil && p.next.item != item {
				p = p.next
			}
			if (p != nil) {
				/* then we must have found the item. */
				/* unlink the item's holding cell... */
				c := p.next				
				p.next = c.next
				/* make sure c isn't the tail */
				if (c == q.tail) {
					q.tail = p
				}
				found = true
			}
		}
	}
	q.l.Unlock()
	return found
}


/* Count the items in the queue */
func (q *Queue) Length() (count int) {
	count = 0

	q.l.RLock()
	for p := q.head; p != nil; p = p.next {
		count += 1
	}
	q.l.RUnlock()

	return count
}

/* Return an iterator for the queue object */
func (q *Queue) Iter() (iter *QueueIter) {
	iter = new(QueueIter)
	iter.q = q
	iter.Reset()

	return iter
}
