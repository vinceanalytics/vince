!function(){"use strict";var o=window.location,l=window.document,u=l.currentScript,s=u.getAttribute("data-api")||new URL(u.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=u&&u.getAttribute("data-include"),n=u&&u.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(r),n=n&&n.split(",").some(r);if(!a||n)return c("exclusion rule",t)}function r(e){var t=o.pathname;return(t+=o.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=e,a.u=o.href,a.d=u.getAttribute("data-domain"),a.r=l.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),a.h=1,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function r(){a=o.pathname,e("pageview")}function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",r),"prerender"===l.visibilityState?l.addEventListener("visibilitychange",function(){a||"visible"!==l.visibilityState||r()}):r();var f=1;function i(e){var t,a;if("auxclick"!==e.type||e.button===f)return(t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||m<a)return!1;if(h(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(a=t)&&a.href&&a.host&&a.host!==o.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function d(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}function v(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}l.addEventListener("click",i),l.addEventListener("auxclick",i);var m=3;function g(e){if("auxclick"!==e.type||e.button===f){for(var t,a,n,r,i=e.target,o=0;o<=m&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;p(i)&&(t=i),h(i)&&(a=i),i=i.parentNode}a&&(r=v(a),t?(r.props.url=t.href,d(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}l.addEventListener("submit",function(e){var t,a=e.target,n=v(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),l.addEventListener("click",g),l.addEventListener("auxclick",g)}();