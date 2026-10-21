package op

import (
	"fmt"
	stdpath "path"
	"sort"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
)

// PendingDebugSummary returns up to limit pending-add entries of the directory
// that contains actualPath, formatted as "name(age=..)". Diagnostic only.
func PendingDebugSummary(storage driver.Driver, actualPath string, limit int) []string {
	if storage == nil || storage.Config().NoCache {
		return nil
	}
	dir, _ := stdpath.Split(actualPath)
	dirKey := Key(storage, dir)
	now := time.Now()
	pendings.mu.Lock()
	d, ok := pendings.dirs[dirKey]
	type ent struct {
		name string
		age  time.Duration
	}
	var ents []ent
	if ok {
		for n, e := range d.add {
			if now.After(e.expiresAt) {
				continue
			}
			ents = append(ents, ent{n, pendingTTL - e.expiresAt.Sub(now)})
		}
	}
	pendings.mu.Unlock()
	sort.Slice(ents, func(i, j int) bool { return ents[i].age < ents[j].age })
	if len(ents) > limit {
		ents = ents[:limit]
	}
	out := make([]string, 0, len(ents))
	for _, e := range ents {
		out = append(out, fmt.Sprintf("%q(age=%s)", e.name, e.age.Truncate(time.Millisecond)))
	}
	return out
}
