#!/bin/bash -e

if [ -z "${VERSION}" ]; then
	echo "You must specify a tag to release: VERSION=1.0.0 make release"
	exit 1
fi

if [ ! -z "$(git tag -l ${VERSION})" ]; then
	echo "Tag ${VERSION} already exists"
	exit 1
fi

if [[ $(git branch --show-current) != "main" ]]; then
	echo "You must run this target on main branch"
	exit 1
fi

if [ ! -z "$(git status --short)" ]; then
	echo "You can't have pending changes when running this target, please stash or push any changes"
	exit 1
fi

if [ ! -z "$(git fetch --dry-run)" ]; then
	echo "Your local main branch is not up-to-date with the remote main branch, please pull"
	exit 1
fi

echo "Generating install manifest..."
helm template ./chart/ --set images.tag=${VERSION} --set images.controller=datadog/chaos-controller --set images.injector=datadog/chaos-injector --set images.handler=datadog/chaos-handler > ./chart/install.yaml
git add ./chart/install.yaml
git commit -m "Generate install manifest for version ${VERSION}"
echo "Creating git tag..."
git tag -a ${VERSION} -m "Release ${VERSION}"
echo "All done! Please run the following command when you feel ready:"
echo "\t --> git push origin main --follow-tags <--"
