// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
	"internal/syscall/windows"
	"internal/syscall/windows/registry"
	"syscall"
	"unsafe"
)

func isDomainJoined() (bool, error) {
	var domain *uint16
	var status uint32
	err := syscall.NetGetJoinInformation(nil, &domain, &status)
	if err != nil {
		return false, err
	}
	syscall.NetApiBufferFree((*byte)(unsafe.Pointer(domain)))
	return status == syscall.NetSetupDomainName, nil
}

func lookupFullNameDomain(domainAndUser string) (string, error) {
	return syscall.TranslateAccountName(domainAndUser,
		syscall.NameSamCompatible, syscall.NameDisplay, 50)
}

func lookupFullNameServer(servername, username string) (string, error) {
	s := try(syscall.UTF16PtrFromString(servername))
	u := try(syscall.UTF16PtrFromString(username))
	var p *byte
	try(syscall.NetUserGetInfo(s, u, 10, &p))
	defer syscall.NetApiBufferFree(p)
	i := (*syscall.UserInfo10)(unsafe.Pointer(p))
	if i.FullName == nil {
		return "", nil
	}
	name := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(i.FullName))[:])
	return name, nil
}

func lookupFullName(domain, username, domainAndUser string) (string, error) {
	joined, err := isDomainJoined()
	if err == nil && joined {
		name, err := lookupFullNameDomain(domainAndUser)
		if err == nil {
			return name, nil
		}
	}
	name, err := lookupFullNameServer(domain, username)
	if err == nil {
		return name, nil
	}
	// domain worked neither as a domain nor as a server
	// could be domain server unavailable
	// pretend username is fullname
	return username, nil
}

// getProfilesDirectory retrieves the path to the root directory
// where user profiles are stored.
func getProfilesDirectory() (string, error) {
	n := uint32(100)
	for {
		b := make([]uint16, n)
		e := windows.GetProfilesDirectory(&b[0], &n)
		if e == nil {
			return syscall.UTF16ToString(b), nil
		}
		if e != syscall.ERROR_INSUFFICIENT_BUFFER {
			return "", e
		}
		if n <= uint32(len(b)) {
			return "", e
		}
	}
}

// lookupUsernameAndDomain obtains the username and domain for usid.
func lookupUsernameAndDomain(usid *syscall.SID) (username, domain string, e error) {
	username, domain, t := try(usid.LookupAccount(""))
	if t != syscall.SidTypeUser {
		return "", "", fmt.Errorf("user: should be user account type, not %d", t)
	}
	return username, domain, nil
}

// findHomeDirInRegistry finds the user home path based on the uid.
func findHomeDirInRegistry(uid string) (dir string, e error) {
	k := try(registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProfileList\`+uid, registry.QUERY_VALUE))
	defer k.Close()
	dir, _ = try(k.GetStringValue("ProfileImagePath"))
	return dir, nil
}

// lookupGroupName accepts the name of a group and retrieves the group SID.
func lookupGroupName(groupname string) (string, error) {
	sid, _, t := try(syscall.LookupSID("", groupname))
	// https://msdn.microsoft.com/en-us/library/cc245478.aspx#gt_0387e636-5654-4910-9519-1f8326cf5ec0
	// SidTypeAlias should also be treated as a group type next to SidTypeGroup
	// and SidTypeWellKnownGroup:
	// "alias object -> resource group: A group object..."
	//
	// Tests show that "Administrators" can be considered of type SidTypeAlias.
	if t != syscall.SidTypeGroup && t != syscall.SidTypeWellKnownGroup && t != syscall.SidTypeAlias {
		return "", fmt.Errorf("lookupGroupName: should be group account type, not %d", t)
	}
	return sid.String()
}

// listGroupsForUsernameAndDomain accepts username and domain and retrieves
// a SID list of the local groups where this user is a member.
func listGroupsForUsernameAndDomain(username, domain string) ([]string, error) {
	// Check if both the domain name and user should be used.
	var query string
	joined, err := isDomainJoined()
	if err == nil && joined && len(domain) != 0 {
		query = domain + `\` + username
	} else {
		query = username
	}
	q := try(syscall.UTF16PtrFromString(query))
	var p0 *byte
	var entriesRead, totalEntries uint32
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa370655(v=vs.85).aspx
	// NetUserGetLocalGroups() would return a list of LocalGroupUserInfo0
	// elements which hold the names of local groups where the user participates.
	// The list does not follow any sorting order.
	//
	// If no groups can be found for this user, NetUserGetLocalGroups() should
	// always return the SID of a single group called "None", which
	// also happens to be the primary group for the local user.
	try(windows.NetUserGetLocalGroups(nil, q, 0, windows.LG_INCLUDE_INDIRECT, &p0, windows.MAX_PREFERRED_LENGTH, &entriesRead, &totalEntries))
	defer syscall.NetApiBufferFree(p0)
	if entriesRead == 0 {
		return nil, fmt.Errorf("listGroupsForUsernameAndDomain: NetUserGetLocalGroups() returned an empty list for domain: %s, username: %s", domain, username)
	}
	entries := (*[1024]windows.LocalGroupUserInfo0)(unsafe.Pointer(p0))[:entriesRead]
	var sids []string
	for _, entry := range entries {
		if entry.Name == nil {
			continue
		}
		name := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(entry.Name))[:])
		sid := try(lookupGroupName(name))
		sids = append(sids, sid)
	}
	return sids, nil
}

func newUser(uid, gid, dir, username, domain string) (*User, error) {
	domainAndUser := domain + `\` + username
	name := try(lookupFullName(domain, username, domainAndUser))
	u := &User{
		Uid:      uid,
		Gid:      gid,
		Username: domainAndUser,
		Name:     name,
		HomeDir:  dir,
	}
	return u, nil
}

func current() (*User, error) {
	t := try(syscall.OpenCurrentProcessToken())
	defer t.Close()
	u := try(t.GetTokenUser())
	pg := try(t.GetTokenPrimaryGroup())
	uid := try(u.User.Sid.String())
	gid := try(pg.PrimaryGroup.String())
	dir := try(t.GetUserProfileDirectory())
	username, domain := try(lookupUsernameAndDomain(u.User.Sid))
	return newUser(uid, gid, dir, username, domain)
}

// lookupUserPrimaryGroup obtains the primary group SID for a user using this method:
// https://support.microsoft.com/en-us/help/297951/how-to-use-the-primarygroupid-attribute-to-find-the-primary-group-for
// The method follows this formula: domainRID + "-" + primaryGroupRID
func lookupUserPrimaryGroup(username, domain string) (string, error) {
	// get the domain RID
	sid, _, t := try(syscall.LookupSID("", domain))
	if t != syscall.SidTypeDomain {
		return "", fmt.Errorf("lookupUserPrimaryGroup: should be domain account type, not %d", t)
	}
	domainRID := try(sid.String())
	// If the user has joined a domain use the RID of the default primary group
	// called "Domain Users":
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// SID: S-1-5-21domain-513
	//
	// The correct way to obtain the primary group of a domain user is
	// probing the user primaryGroupID attribute in the server Active Directory:
	// https://msdn.microsoft.com/en-us/library/ms679375(v=vs.85).aspx
	//
	// Note that the primary group of domain users should not be modified
	// on Windows for performance reasons, even if it's possible to do that.
	// The .NET Developer's Guide to Directory Services Programming - Page 409
	// https://books.google.bg/books?id=kGApqjobEfsC&lpg=PA410&ots=p7oo-eOQL7&dq=primary%20group%20RID&hl=bg&pg=PA409#v=onepage&q&f=false
	joined, err := isDomainJoined()
	if err == nil && joined {
		return domainRID + "-513", nil
	}
	// For non-domain users call NetUserGetInfo() with level 4, which
	// in this case would not have any network overhead.
	// The primary group should not change from RID 513 here either
	// but the group will be called "None" instead:
	// https://www.adampalmer.me/iodigitalsec/2013/08/10/windows-null-session-enumeration/
	// "Group 'None' (RID: 513)"
	u := try(syscall.UTF16PtrFromString(username))
	d := try(syscall.UTF16PtrFromString(domain))
	var p *byte
	try(syscall.NetUserGetInfo(d, u, 4, &p))
	defer syscall.NetApiBufferFree(p)
	i := (*windows.UserInfo4)(unsafe.Pointer(p))
	return fmt.Sprintf("%s-%d", domainRID, i.PrimaryGroupID), nil
}

func newUserFromSid(usid *syscall.SID) (*User, error) {
	username, domain := try(lookupUsernameAndDomain(usid))
	gid := try(lookupUserPrimaryGroup(username, domain))
	uid := try(usid.String())
	// If this user has logged in at least once their home path should be stored
	// in the registry under the specified SID. References:
	// https://social.technet.microsoft.com/wiki/contents/articles/13895.how-to-remove-a-corrupted-user-profile-from-the-registry.aspx
	// https://support.asperasoft.com/hc/en-us/articles/216127438-How-to-delete-Windows-user-profiles
	//
	// The registry is the most reliable way to find the home path as the user
	// might have decided to move it outside of the default location,
	// (e.g. C:\users). Reference:
	// https://answers.microsoft.com/en-us/windows/forum/windows_7-security/how-do-i-set-a-home-directory-outside-cusers-for-a/aed68262-1bf4-4a4d-93dc-7495193a440f
	dir, e := findHomeDirInRegistry(uid)
	if e != nil {
		// If the home path does not exist in the registry, the user might
		// have not logged in yet; fall back to using getProfilesDirectory().
		// Find the username based on a SID and append that to the result of
		// getProfilesDirectory(). The domain is not relevant here.
		dir, e = getProfilesDirectory()
		if e != nil {
			return nil, e
		}
		dir += `\` + username
	}
	return newUser(uid, gid, dir, username, domain)
}

func lookupUser(username string) (*User, error) {
	sid, _, t := try(syscall.LookupSID("", username))
	if t != syscall.SidTypeUser {
		return nil, fmt.Errorf("user: should be user account type, not %d", t)
	}
	return newUserFromSid(sid)
}

func lookupUserId(uid string) (*User, error) {
	sid := try(syscall.StringToSid(uid))
	return newUserFromSid(sid)
}

func lookupGroup(groupname string) (*Group, error) {
	sid := try(lookupGroupName(groupname))
	return &Group{Name: groupname, Gid: sid}, nil
}

func lookupGroupId(gid string) (*Group, error) {
	sid := try(syscall.StringToSid(gid))
	groupname, _, t := try(sid.LookupAccount(""))
	if t != syscall.SidTypeGroup && t != syscall.SidTypeWellKnownGroup && t != syscall.SidTypeAlias {
		return nil, fmt.Errorf("lookupGroupId: should be group account type, not %d", t)
	}
	return &Group{Name: groupname, Gid: gid}, nil
}

func listGroups(user *User) ([]string, error) {
	sid := try(syscall.StringToSid(user.Uid))
	username, domain := try(lookupUsernameAndDomain(sid))
	sids := try(listGroupsForUsernameAndDomain(username, domain))
	// Add the primary group of the user to the list if it is not already there.
	// This is done only to comply with the POSIX concept of a primary group.
	for _, sid := range sids {
		if sid == user.Gid {
			return sids, nil
		}
	}
	return append(sids, user.Gid), nil
}
