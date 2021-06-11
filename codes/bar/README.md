# prepare environment
make prepapre-env

# deploy
make build image deploy

# check service status
make dinfo

# clear deployment
make undeploy

# to test if bar app works locally
make testapp
