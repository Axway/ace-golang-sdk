#!groovy
pipeline {
    agent {
        docker {
            image 'ace.swf-artifactory.lab.phx.axway.int/ace/alpine-golang-librdkafka:latest'
        }
    }

    parameters {
        /* Parameter specifications below are generic, they are shared by all microservices of a same deployment */
        string(
            name: 'SONAR_URL',
            description: 'URL for Sonar',
            defaultValue: params.SONAR_URL ?: '')
        string(
            name: 'SONAR_PROJECT_NAME',
            description: 'Name of the sonar project',
            defaultValue: params.SONAR_PROJECT_NAME ?: 'ACE-SDK-Go')
        string(
            name: 'SONAR_PROJECT_VERSION',
            description: 'Version of the sonar project',
            defaultValue: params.SONAR_PROJECT_VERSION ?: '')
    }
    stages {
        stage('ACE Golang SDK Test and Sonar Scan') {
            steps {
                sh '''
                    mkdir -p $GOPATH/src/github.com/Axway
                    cd $GOPATH/src/github.com/Axway
                    ln -s ${WORKSPACE} ace-golang-sdk
                    cd $GOPATH/src/github.com/Axway/ace-golang-sdk
                    go test -v -short -coverpkg=./... -coverprofile=$GOPATH/src/github.com/Axway/ace-golang-sdk/gocoverage.out -count=1 ./...
                    sonar-scanner -X \
                        -Dsonar.host.url=${SONAR_URL} \
                        -Dsonar.language=go \
                        -Dsonar.projectName=${SONAR_PROJECT_NAME} \
                        -Dsonar.projectVersion=${SONAR_PROJECT_VERSION} \
                        -Dsonar.projectKey=${SONAR_PROJECT_NAME} \
                        -Dsonar.sourceEncoding=UTF-8 \
                        -Dsonar.projectBaseDir=${GOPATH}/src/github.com/Axway/ace-golang-sdk \
                        -Dsonar.sources=. \
                        -Dsonar.tests=. \
                        -Dsonar.exclusions=**/messaging/*.go,**/rpc/*.go,**/vendor/** \
                        -Dsonar.coverage.exclusions=**/messaging/*.go,**/rpc/*.go,**/vendor/** \
                        -Dsonar.test.inclusions=**/*test*.go \
                        -Dsonar.go.tests.reportPaths=$GOPATH/src/github.com/Axway/ace-golang-sdk/goreport.json \
                        -Dsonar.go.coverage.reportPaths=$GOPATH/src/github.com/Axway/ace-golang-sdk/gocoverage.out
                '''
            }
        }
    }
}
